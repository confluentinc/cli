#!/usr/bin/env bash

# Copyright 2017 Confluent Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

success=false
# Uncomment to enable debugging on the console
# set -x

#currently internal option only.
_default_curl_timeout=10

platform="$( uname -s )"

ERROR_CODE=127

# Exit with an error message.
die() {
    echo "$@"
    exit ${ERROR_CODE}
}

validate_and_export_dir_layout() {
    # We don't need to know CONFLUENT_HOME just to see the list of commands
    [[ $# -lt 1 ]] || [[ "$1" = "skip" ]] && command_name="confluent local" && return

    if [[ -z "${CONFLUENT_HOME}" ]]; then
        command_name="$( basename "${BASH_SOURCE[0]}" )"
        confluent_bin="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
        confluent_home="$( dirname "${confluent_bin}" )"
    else
        command_name="confluent local"
        confluent_home="${CONFLUENT_HOME}"
        confluent_bin="${CONFLUENT_HOME}/bin"
    fi

    confluent_conf="${confluent_home}/etc"
    # workaround for the cases when 'etc' is not under the same directory as 'bin'
    if [[ ! -f "${confluent_conf}/schema-registry/connect-avro-distributed.properties" ]]; then
        confluent_conf="$( cd "${confluent_home}/../etc" > /dev/null 2>&1 && pwd )"
    fi

    [[ ! -f "${confluent_conf}/schema-registry/connect-avro-distributed.properties" ]] \
        && die "Cannot locate 'etc' directory for Confluent Platform."

    # $TMPDIR includes a trailing '/' by default.
    tmp_dir="${TMPDIR:-/tmp/}"
    confluent_current_dir="${CONFLUENT_CURRENT:-${tmp_dir}}"
    last="${confluent_current_dir:${#confluent_current_dir}-1:1}"
    [[ "${last}" != "/" ]] && export confluent_current_dir="${confluent_current_dir}/"
}

# Contains the result of functions that intend to return a value besides their exit status.
_retval=""

Gre='\e[0;32m';
Red='\e[0;31m';
#Reset color
RC='\e[0m'


# Explicit & Hardcoded dependencies for starting and stopping
# Striking the balance between convenience and full control in
# most common use cases.
declare -a deps_start_zookeeper=(
)
declare -a deps_stop_zookeeper=(
    "control-center"
    "ksql-server"
    "connect"
    "kafka-rest"
    "schema-registry"
    "kafka"
)

declare -a deps_start_kafka=(
    "zookeeper"
)
declare -a deps_stop_kafka=(
    "control-center"
    "ksql-server"
    "connect"
    "kafka-rest"
    "schema-registry"
)

declare -a deps_start_schema_registry=(
    "zookeeper"
    "kafka"
)
declare -a deps_stop_schema_registry=(
)

declare -a deps_start_kafka_rest=(
    "zookeeper"
    "kafka"
    "schema-registry"
)
declare -a deps_stop_kafka_rest=(
)

declare -a deps_start_connect=(
    "zookeeper"
    "kafka"
    "schema-registry"
)
declare -a deps_stop_connect=(
)

declare -a deps_start_ksql_server=(
    "zookeeper"
    "kafka"
    "schema-registry"
)
declare -a deps_stop_ksql_server=(
)

declare -a deps_start_control_center=(
    "zookeeper"
    "kafka"
)
declare -a deps_stop_control_center=(
)

declare -a services=(
    "zookeeper"
    "kafka"
    "schema-registry"
    "kafka-rest"
    "connect"
    "ksql-server"
)

declare -a rev_services=(
    "ksql-server"
    "connect"
    "kafka-rest"
    "schema-registry"
    "kafka"
    "zookeeper"
)

declare -a confluent_platform_services=(
    "control-center"
)

declare -a rev_confluent_platform_services=(
    "control-center"
)

declare -a commands=(
    "list"
    "start"
    "stop"
    "status"
    "current"
    "destroy"
    "demo"
    "top"
    "log"
    "load"
    "unload"
    "produce"
    "consume"
    "config"
    "version"
)

declare -a confluent_platform_commands=(
    "acl"
)

declare -a connector_properties=(
    "elasticsearch-sink=kafka-connect-elasticsearch/quickstart-elasticsearch.properties"
    "file-source=kafka/connect-file-source.properties"
    "file-sink=kafka/connect-file-sink.properties"
    "jdbc-source=kafka-connect-jdbc/source-quickstart-sqlite.properties"
    "jdbc-sink=kafka-connect-jdbc/sink-quickstart-sqlite.properties"
    "hdfs-sink=kafka-connect-hdfs/quickstart-hdfs.properties"
    "s3-sink=kafka-connect-s3/quickstart-s3.properties"
)

export SAVED_KAFKA_LOG4J_OPTS="${KAFKA_LOG4J_OPTS}"
export SAVED_KAFKA_EXTRA_ARGS="${KAFKA_EXTRA_ARGS}"

export SAVED_KAFKA_HEAP_OPTS="${KAFKA_HEAP_OPTS}"
export SAVED_KAFKA_JVM_PERFORMANCE_OPTS="${KAFKA_JVM_PERFORMANCE_OPTS}"
export SAVED_KAFKA_GC_LOG_OPTS="${KAFKA_GC_LOG_OPTS}"

export SAVED_KAFKA_JMX_OPTS="${KAFKA_JMX_OPTS}"
export SAVED_KAFKA_JMX_PORT="${KAFKA_JMX_PORT}"

export SAVED_KAFKA_DEBUG="${KAFKA_DEBUG}"
export SAVED_KAFKA_OPTS="${KAFKA_OPTS}"

export SAVED_KAFKA_CLASSPATH="${KAFKA_CLASSPATH}"

FORMAT_CMD="jq '.'"

# Which java to use
if [[ "x${JAVA_HOME}" == "x" ]]; then
  JAVA="java"
else
  JAVA="${JAVA_HOME}/bin/java"
fi

is_command() {
  command -v "$1" >/dev/null
}

requirements() {
    local major=3
    local minor=2
    [ "${BASH_VERSINFO[0]:-0}" -lt "${major}" ] \
        || [ "${BASH_VERSINFO[0]:-0}" -eq ${major} -a "${BASH_VERSINFO[1]:-0}" -lt ${minor} ] \
        && invalid_requirement "bash" "${major}.${minor}"

    is_command curl || invalid_requirement "curl"
    is_command jq || FORMAT_CMD="xargs -0"
}

export_service_env() {
    # The prefix needs to include any delimiters (e.g. '_').
    local prefix="${1}"

    local var="${prefix}LOG4J_OPTS"
    export KAFKA_LOG4J_OPTS="${!var}"
    var="${prefix}EXTRA_ARGS"
    export EXTRA_ARGS="${!var}"

    var="${prefix}HEAP_OPTS"
    export KAFKA_HEAP_OPTS="${!var}"
    var="${prefix}JVM_PERFORMANCE_OPTS"
    export KAFKA_JVM_PERFORMANCE_OPTS="${!var}"
    var="${prefix}GC_LOG_OPTS"
    export KAFKA_GC_LOG_OPTS="${!var}"

    var="${prefix}JMX_OPTS"
    export KAFKA_JMX_OPTS="${!var}"

    var="${prefix}DEBUG"
    export KAFKA_DEBUG="${!var}"
    var="${prefix}OPTS"
    export KAFKA_OPTS="${!var}"

    var="${prefix}CLASSPATH"
    export CLASSPATH="${!var}"

    var="${prefix}JMX_PORT"
    export JMX_PORT="${!var}"
}

echo_variable() {
    local var_value="${!1}"
    echo "${1} = ${var_value}"
}

# Implies zero or positive
is_integer() {
    [[ -n "${1}" ]] && [[ "${1}" =~ ^[0-9]+$ ]]
}

is_json() {
    which jq > /dev/null 2>&1
    status=$?
    if [[ ${status} -ne 0 ]]; then
        echo "Warning: Install 'jq' to add support for parsing JSON"
        local ext="${config_file##*.}"
        if [[ "${#ext}" -ne 4 ]]; then
            return ${ERROR_CODE}
        fi
        echo "${ext}" | grep -i json > /dev/null 2>&1
        return $?
    fi

    # Check whether we have json contents.
    cat "${config_file}" | eval ${FORMAT_CMD} > /dev/null 2>&1
    return $?
}

# Will pick the last integer value that will encounter
get_service_port() {
    local property="${1}"
    local config_file="${2}"
    local delim="${3:-:}"

    [[ -z "${property}" ]] && die "Need property key to extract service port from configuration"
    local property_split=( $( grep -i "^${property}" "${config_file}" | tr "${delim}" "\n" ) )

    _retval=""
    local entry=""
    for entry in "${property_split[@]}"; do
        # trim string
        entry=$( echo "${entry}" | xargs )
        is_integer "${entry}" && _retval="${entry}"
    done
}

wheel_pos=0
wheel_freq_ms=100
spinner_running=false

spinner_init() {
    spinner_running=false
}

spinner() {
    local wheel='-\|/'
    wheel_pos=$(( (wheel_pos + 1) % ${#wheel} ))
    printf "\r${wheel:${wheel_pos}:1}"
    spinner_running=true
    sleep 0.${wheel_freq_ms}
}

spinner_done() {
    # Backspace to override spinner in the next printf/echo
    [[ "${spinner_running}" == true ]] && printf "\b"
    spinner_running=false
}

is_confluent_platform() {
    local confluent_platform_prefix="${confluent_home}/share/java/confluent-control-center/control-center-"
    confluent_version="$( ls ${confluent_platform_prefix}*.jar 2> /dev/null )"
    status=$?
    return ${status}
}

get_version() {
    local confluent_platform_prefix="${confluent_home}/share/java/kafka-connect-replicator/kafka-connect-replicator-"
    local ccs_prefix="${confluent_home}/share/java/confluent-common/common-config-"
    local kafka_prefix="${confluent_home}/share/java/kafka/kafka-clients-"
    local zookeeper_prefix="${confluent_home}/share/java/kafka/zookeeper-"

    confluent_version="$( ls ${confluent_platform_prefix}*.jar 2> /dev/null )"
    status=$?
    if [[ ${status} -eq 0 ]]; then
        export confluent_flavor="Confluent Platform"
        confluent_version="${confluent_version#$confluent_platform_prefix}"
        export confluent_version="${confluent_version%.jar}"
    else
        confluent_version="$( ls ${ccs_prefix}*.jar 2> /dev/null )"
        export confluent_flavor="Confluent Community Software"
        confluent_version="${confluent_version#$ccs_prefix}"
        export confluent_version="${confluent_version%.jar}"
    fi

    kafka_version="$( ls ${kafka_prefix}*.jar 2> /dev/null )"
    kafka_version="${kafka_version#$kafka_prefix}"
    export kafka_version="${kafka_version%.jar}"

    zookeeper_version="$( ls ${zookeeper_prefix}*.jar 2> /dev/null )"
    zookeeper_version="${zookeeper_version#$zookeeper_prefix}"
    export zookeeper_version="${zookeeper_version%.jar}"
}

set_or_get_current() {
    if [[ -f "${confluent_current_dir}confluent.current" ]]; then
        export confluent_current="$( cat "${confluent_current_dir}confluent.current" )"
    fi

    if [[ ! -d "${confluent_current}" ]]; then
        export confluent_current="$( mktemp -d ${confluent_current_dir}confluent.XXXXXXXX )"
        echo "${confluent_current}" > "${confluent_current_dir}confluent.current"
    fi

    get_version
}

shutdown() {
    [[ "${success}" == false ]] \
        && echo "Unsuccessful execution. Attempting service shutdown" \
        && stop_command

}

is_alive() {
    local pid="${1}"
    kill -0 "${pid}" > /dev/null 2>&1
}

is_not_alive() {
    ! is_alive "${1}"
}

is_running() {
    local service="${1}"
    local print_status="${2}"
    local service_dir="${confluent_current}/${service}"
    local service_pid="$( cat "${service_dir}/${service}.pid" 2> /dev/null )"

    is_alive "${service_pid}"
    status=$?
    if [[ "${print_status}" != false ]]; then
        if [[ ${status} -eq 0 ]]; then
            printf "${service} is [${Gre}UP${RC}]\n"
        else
            printf "${service} is [${Red}DOWN${RC}]\n"
        fi
    fi

    return ${status}
}

wait_process_up() {
    #TODO: need to add port condition here too probably
    wait_process "${1}" "up" "${2}"
}

wait_process_down() {
    wait_process "${1}" "down" "${2}"
}

wait_process() {
    local pid="${1}"
    local event="${2:-up}"
    local timeout_ms="${3}"

    is_integer "${pid}" || die "Need PID to wait on a process"

    # Default max wait time set to 10 minutes. That's practically infinite for this program.
    is_integer "${timeout_ms}" || timeout_ms=600000

    local mode=is_alive
    spinner_init
    # Busy wait in case service dies soon after startup
    while ${mode} "${pid}" && [[ "${timeout_ms}" -gt 0 ]]; do
        spinner
        (( timeout_ms = timeout_ms - wheel_freq_ms ))
    done
    spinner_done

    if [[ "${event}" == "down" ]]; then
        ! ${mode} "${pid}"
    else
        ${mode} "${pid}"
    fi
}

stop_and_wait_process() {
    local pid="${1}"
    local timeout_ms="${2}"
    # Default max wait time set to 10 minutes. That's practically infinite for this program.
    is_integer "${timeout_ms}" || timeout_ms=600000

    spinner_init
    kill "${pid}" 2> /dev/null
    while kill -0 "${pid}" > /dev/null 2>&1 && [[ "${timeout_ms}" -gt 0 ]]; do
        spinner
        (( timeout_ms = timeout_ms - wheel_freq_ms ))
    done
    spinner_done
    # Will have no effect if the process stopped gracefully
    # TODO: maybe should issue a warning if the process is not stopped gracefully.
    kill -9 "${pid}" > /dev/null 2>&1
}

start_zookeeper() {
    export_service_env "ZOOKEEPER_"
    export_log4j_zookeeper
    start_service "zookeeper" "${confluent_bin}/zookeeper-server-start"
}

config_zookeeper() {
    config_service "zookeeper" "kafka" "zookeeper" "dataDir"
}

export_zookeeper() {
    get_service_port "clientPort" "${confluent_conf}/kafka/zookeeper.properties" "="
    if [[ -n "${_retval}" ]]; then
        export zk_port="${_retval}"
    else
        export zk_port="2181"
    fi
}

export_log4j_zookeeper() {
    export_log4j_with_generic_log_dir "zookeeper"
}

stop_zookeeper() {
    stop_service "zookeeper"
}

#TODO: a generic wait_service function makes sense after all.
wait_zookeeper() {
    local pid="${1}"
    export_zookeeper

    local started=false
    local timeout_ms=5000
    while [[ "${started}" == false && "${timeout_ms}" -gt 0 ]]; do
        ( lsof -P -c java 2> /dev/null | grep ${zk_port} > /dev/null 2>&1 ) && started=true
        spinner && (( timeout_ms = timeout_ms - wheel_freq_ms ))
    done
    wait_process_up "${pid}" 2000 || echo "Zookeeper failed to start"
}

start_kafka() {
    local service="kafka"
    is_running "zookeeper" "false" \
        || die "Cannot start Kafka, Zookeeper is not running. Check your deployment"
    export_service_env "SAVED_KAFKA_"
    export_log4j_kafka
    start_service "kafka" "${confluent_bin}/kafka-server-start"
}

config_kafka() {
    export_kafka
    config_service "kafka" "kafka" "server" "log.dirs"
    enable_metrics_reporter "kafka"
}

export_kafka() {
    get_service_port "listeners" "${confluent_conf}/kafka/server.properties"
    if [[ -n "${_retval}" ]]; then
        export kafka_port="${_retval}"
    else
        export kafka_port="9092"
    fi
}

export_log4j_kafka() {
    export_log4j_with_generic_log_dir "kafka"
}

stop_kafka() {
    stop_service "kafka"
}

wait_kafka() {
    local pid="${1}"
    export_kafka

    local started=false
    local timeout_ms=10000

    while [[ "${started}" == false && "${timeout_ms}" -gt 0 ]]; do
        ( lsof -P -c java 2> /dev/null | grep ${kafka_port} > /dev/null 2>&1 ) && started=true
        spinner && (( timeout_ms = timeout_ms - wheel_freq_ms ))
    done
    wait_process_up "${pid}" 3000 || echo "Kafka failed to start"
}

start_schema-registry() {
    local service="schema-registry"
    is_running "kafka" "false" \
        || die "Cannot start Schema Registry, Kafka Server is not running. Check your deployment"
    export_service_env "SCHEMA_REGISTRY_"
    export_log4j_schema-registry
    start_service "schema-registry" "${confluent_bin}/schema-registry-start"
}

config_schema-registry() {
    export_zookeeper
    config_service "schema-registry" "schema-registry" "schema-registry"\
        "kafkastore.connection.url" "localhost:${zk_port}"
    enable_monitoring_interceptors "schema-registry"
}

export_schema-registry() {
    get_service_port "listeners" "${confluent_conf}/schema-registry/schema-registry.properties"
    if [[ -n "${_retval}" ]]; then
        export schema_registry_port="${_retval}"
    else
        export schema_registry_port="8081"
    fi
}

export_log4j_schema-registry() {
    export_log4j_with_generic_log_dir "schema-registry"
}

stop_schema-registry() {
    stop_service "schema-registry"
}

wait_schema-registry() {
    local pid="${1}"
    export_schema-registry

    local started=false
    local timeout_ms=5000
    while [[ "${started}" == false && "${timeout_ms}" -gt 0 ]]; do
        ( lsof -P -c java 2> /dev/null | grep ${schema_registry_port} > /dev/null 2>&1 ) && started=true
        spinner && (( timeout_ms = timeout_ms - wheel_freq_ms ))
    done
    wait_process_up "${pid}" 2000 || echo "Schema Registry failed to start"
}

start_kafka-rest() {
    local service="kafka-rest"
    is_running "kafka" "false" \
        || die "Cannot start Kafka Rest, Kafka Server is not running. Check your deployment"
    export_service_env "KAFKAREST_"
    export_log4j_kafka-rest
    start_service "kafka-rest" "${confluent_bin}/kafka-rest-start"
}

config_kafka-rest() {
    export_zookeeper
    export_schema-registry

    config_service "kafka-rest" "kafka-rest" "kafka-rest"\
        "zookeeper.connect" "localhost:${zk_port}"

    config_service "kafka-rest" "kafka-rest" "kafka-rest"\
        "schema.registry.url" "http://localhost:${schema_registry_port}" "reapply"

    enable_monitoring_interceptors "kafka-rest"
}

export_kafka-rest() {
    get_service_port "listeners" "${confluent_conf}/kafka-rest/kafka-rest.properties"
    if [[ -n "${_retval}" ]]; then
        export kafka_rest_port="${_retval}"
    else
        export kafka_rest_port="8082"
    fi
}

export_log4j_kafka-rest() {
    export_log4j_with_generic_log_dir "kafka-rest"
}

stop_kafka-rest() {
    stop_service "kafka-rest"
}

wait_kafka-rest() {
    local pid="${1}"
    export_kafka-rest

    local started=false
    local timeout_ms=5000
    while [[ "${started}" == false && "${timeout_ms}" -gt 0 ]]; do
        ( lsof -P -c java 2> /dev/null | grep ${kafka_rest_port} > /dev/null 2>&1 ) && started=true
        spinner && (( timeout_ms = timeout_ms - wheel_freq_ms ))
    done
    wait_process_up "${pid}" 2000 || echo "Kafka Rest failed to start"
}

start_connect() {
    local service="connect"
    is_running "kafka" "false" \
        || die "Cannot start Kafka Connect, Kafka Server is not running. Check your deployment"
    export_service_env "CONNECT_"
    export_log4j_connect
    start_service "connect" "${confluent_bin}/connect-distributed"
}

config_connect() {
    get_service_port "listeners" "${confluent_conf}/kafka/server.properties"
    export_kafka

    config_service "connect" "schema-registry" "connect-avro-distributed" \
        "bootstrap.servers" "localhost:${kafka_port}"

    enable_monitoring_interceptors "connect"
}

export_connect() {
    get_service_port "rest.port" "${confluent_conf}/schema-registry/connect-avro-distributed.properties" "="
    if [[ -n "${_retval}" ]]; then
        export connect_port="${_retval}"
    else
        export connect_port="8083"
    fi
}

export_log4j_connect() {
    export_log4j_with_generic_log_dir "connect"
}

stop_connect() {
    stop_service "connect"
}

wait_connect() {
    local pid="${1}"
    export_connect

    local started=false
    local timeout_ms=10000
    while [[ "${started}" == false && "${timeout_ms}" -gt 0 ]]; do
        ( lsof -P -c java 2> /dev/null | grep ${connect_port} > /dev/null 2>&1 ) && started=true
        spinner && (( timeout_ms = timeout_ms - wheel_freq_ms ))
    done
    wait_process_up "${pid}" 4000 || echo "Kafka Connect failed to start"
}

start_ksql-server() {
    local service="ksql-server"
    is_running "schema-registry" "false" \
        || die "Cannot start ksql-server, Kafka Server or Schema Registry Server is not running. Check your deployment"
    export_service_env "KSQL_"
    export_log4j_ksql-server
    start_service "ksql-server" "${confluent_bin}/ksql-server-start"
}

config_ksql-server() {
    export_zookeeper
    config_service "ksql-server" "ksql" "ksql-server"\
        "kafkastore.connection.url" "localhost:${zk_port}"
    enable_monitoring_interceptors "ksql-server"
}

export_ksql-server() {
    get_service_port "listeners" "${confluent_conf}/ksql/ksql-server.properties"
    if [[ -n "${_retval}" ]]; then
        export ksql_port="${_retval}"
    else
        export ksql_port="8088"
    fi
}

export_log4j_ksql-server() {
    export_log4j_with_generic_log_dir "ksql-server"
}

stop_ksql-server() {
    stop_service "ksql-server"
}

wait_ksql-server() {
    local pid="${1}"
    export_ksql-server

    local started=false
    local timeout_ms=5000
    while [[ "${started}" == false && "${timeout_ms}" -gt 0 ]]; do
        ( lsof -P -c java 2> /dev/null | grep ${ksql_port} > /dev/null 2>&1 ) && started=true
        spinner && (( timeout_ms = timeout_ms - wheel_freq_ms ))
    done
    wait_process_up "${pid}" 2000 || echo "ksql-server failed to start"
}

start_control-center() {
    local service="control-center"
    is_running "connect" "false" \
        || die "Cannot start Control-Center, Kafka Connect is not running. Check your deployment"
    export_service_env "CONTROL_CENTER_"
    export_log4j_control-center
    start_service "control-center" "${confluent_bin}/control-center-start"
}

config_control-center() {
    export_zookeeper
    export_kafka
    export_connect
    config_service "control-center" "confluent-control-center" "control-center-dev" "confluent.controlcenter.data.dir"
}

export_control-center() {
    get_service_port "listeners" "${confluent_conf}/confluent-control-center/control-center-dev.properties"
    if [[ -n "${_retval}" ]]; then
        export control_center_port="${_retval}"
    else
        export control_center_port="9021"
    fi
}

export_log4j_control-center() {
    export_log4j_with_generic_log_dir "control-center"
}

stop_control-center() {
    stop_service "control-center"
}

wait_control-center() {
    local pid="${1}"
    export_control-center

    local started=false
    local timeout_ms=10000
    while [[ "${started}" == false && "${timeout_ms}" -gt 0 ]]; do
        ( lsof -P -c java 2> /dev/null | grep ${control_center_port} > /dev/null 2>&1 ) && started=true
        spinner && (( timeout_ms = timeout_ms - wheel_freq_ms ))
    done
    wait_process_up "${pid}" 2000 || echo "control-center failed to start"
}

status_service() {
    local service="${1}"
    if [[ -n "${service}" ]]; then
        ! service_exists "${service}" && die "Unknown service: ${service}"
    fi

    skip=true
    local entry=""
    [[ -z "${service}" ]] && skip=false
    is_confluent_platform
    status=$?
    if [[ ${status} -eq 0 ]]; then
        for entry in "${rev_confluent_platform_services[@]}"; do
            [[ "${entry}" == "${service}" ]] && skip=false;
            [[ "${skip}" == false ]] && is_running "${entry}"
        done
    fi

    for entry in "${rev_services[@]}"; do
        [[ "${entry}" == "${service}" ]] && skip=false;
        [[ "${skip}" == false ]] && is_running "${entry}"
    done
}

start_service() {
    local service="${1}"
    local start_command="${2}"
    local service_dir="${confluent_current}/${service}"
    is_running "${service}" "false" \
        && echo "${service} is already running. Try restarting if needed"\
        && return 0

    mkdir -p "${service_dir}"
    config_"${service}"
    echo "Starting ${service}"
    # TODO: decide whether to persist logs on stdout / stderr between runs.
    ${start_command} "${service_dir}/${service}.properties" \
        &> "${service_dir}/${service}.stdout" &
    echo $! > "${service_dir}/${service}.pid"
    local service_pid="$( cat "${service_dir}/${service}.pid" 2> /dev/null )"
    wait_"${service}" "${service_pid}"
    is_running "${service}"
}

# The first 3 args seem unavoidable right now. 4th is optional
# TODO: refactor to pass property pairs as a map.
config_service() {
    local service="${1}"
    local package="${2}"
    local property_file="${3}"

    ( [[ -z "${service}" ]] || [[ -z "${package}" ]] || [[ -z "${property_file}" ]] ) \
        && die "Missing required configuration properties for service: ${service}"

    local service_dir="${confluent_current}/${service}"
    mkdir -p "${service_dir}/data"
    local property_key="${4}"
    local property_value="${5}"
    if [[ -n "${property_key}" && -z "${property_value}" ]]; then
        config_command="sed -e s@^${property_key}=.*@${property_key}=${service_dir}/data@g"
    else
        #TODO: Generalize how key-value pairs are set. Property key-value pairs are ignored for now.
        config_command=cat
    fi

    local input_file="${confluent_conf}/${package}/${property_file}.properties"

    ${config_command} < "${input_file}" \
        > "${service_dir}/${service}.properties"

    # Override Connect's config, in case this is an unchanged config from a tar.gz or .zip package
    # installation, to make plugin.path work for any "current working directory (cwd)"
    if [[ "${service}" == "connect" ]]; then
        sed "s@^plugin.path=share/java@plugin.path=${confluent_home}/share/java@g" \
            "${service_dir}/${service}.properties" > "${service_dir}/${service}.properties.bak"
        mv -f "${service_dir}/${service}.properties.bak" "${service_dir}/${service}.properties"
    fi

    # Set ksql-server data dir. TODO: refactor when config_service supports general handling of key-value pairs
    if [[ "${service}" == "ksql-server" ]]; then
        mkdir -p "${service_dir}/data/kafka-streams"
        echo "" >> "${service_dir}/${service}.properties"
        echo "state.dir=${service_dir}/data/kafka-streams" >> "${service_dir}/${service}.properties"
    fi
}

export_log4j_with_generic_log_dir() {
    local service="${1}"
    export LOG_DIR="${confluent_current}/${service}/logs"
}

enable_metrics_reporter() {
    is_confluent_platform
    status=$?
    if [[ ${status} -ne 0 ]]; then
        return 1
    fi

    local service="${1}"

    local service_dir="${confluent_current}/${service}"
    echo "" >> "${service_dir}/${service}.properties"
    echo "metric.reporters=io.confluent.metrics.reporter.ConfluentMetricsReporter" \
        >> "${service_dir}/${service}.properties"
    echo "confluent.metrics.reporter.bootstrap.servers=localhost:${kafka_port}" \
        >> "${service_dir}/${service}.properties"
    echo "confluent.metrics.reporter.topic.replicas=1" >> "${service_dir}/${service}.properties"
    echo "" >> "${service_dir}/${service}.properties"

    return 0
}

enable_monitoring_interceptors() {
    is_confluent_platform
    status=$?
    if [[ ${status} -ne 0 ]]; then
        return 1
    fi

    local service="${1}"

    local service_dir="${confluent_current}/${service}"
    echo "" >> "${service_dir}/${service}.properties"
    echo "producer.interceptor.classes=io.confluent.monitoring.clients.interceptor.MonitoringProducerInterceptor" \
        >> "${service_dir}/${service}.properties"
    echo "consumer.interceptor.classes=io.confluent.monitoring.clients.interceptor.MonitoringConsumerInterceptor" \
        >> "${service_dir}/${service}.properties"
    echo "" >> "${service_dir}/${service}.properties"

    return 0
}

stop_service() {
    local service="${1}"
    local service_dir="${confluent_current}/${service}"
    # check file exists, and if not issue warning.
    local service_pid="$( cat "${service_dir}/${service}.pid" 2> /dev/null )"
    echo "Stopping ${service}"

    stop_and_wait_process "${service_pid}" 10000
    rm -f "${service_dir}/${service}.pid"
    is_running "${service}"
}

service_exists() {
    local service="${1}"
    exists "${service}" "services" && return 0
    is_confluent_platform && exists "${service}" "confluent_platform_services" && return 0
    return 1
}

command_exists() {
    local command="${1}"
    exists "${command}" "commands" || exists "${command}" "confluent_platform_commands"
}

exec_cli(){
    exec "${confluent_bin}"/"${1}"  "$@"
}

exists() {
    local arg="${1}"

    case "${2}" in
        "services")
        local list=( "${services[@]}" ) ;;
        "confluent_platform_services")
        local list=( "${confluent_platform_services[@]}" ) ;;
        "commands")
        local list=( "${commands[@]}" ) ;;
        "confluent_platform_commands")
        local list=( "${confluent_platform_commands[@]}" ) ;;
    esac

    local entry=""
    for entry in "${list[@]}"; do
        [[ ${entry} == "${arg}" ]] && return 0;
    done
    return 1
}

list_command() {
    if [[ "x${1}" == "x" ]]; then
        echo "Available services:"
        local service=""
        for service in "${services[@]}"; do
            echo "  ${service}"
        done

        is_confluent_platform
        status=$?
        if [[ ${status} -eq 0 ]]; then
            for service in "${confluent_platform_services[@]}"; do
                echo "  ${service}"
            done
        fi
    else
        connect_subcommands "list" "$@"
    fi
}

# CLI-84: If using Mac, version must be >= 10.13
validate_os_version() {
    if [[ "x${platform}" = "xDarwin" ]]; then
        local mac_version=$(sw_vers -productVersion)
        local mac_version_major=$(echo ${mac_version} | cut -f 1 -d '.')
        local mac_version_minor=$(echo ${mac_version} | cut -f 2 -d '.')
        if [[ ${mac_version_major} -lt 10 || ( ${mac_version_major} -ge 10 && ${mac_version_minor} -lt 13 ) ]]; then
            die "Error: macOS version >= 10.13 is required for all Confluent Platform features to work (detected macOS: ${mac_version})."
        fi
    fi
}

validate_java_version() {
    local target_service=${1}

    local warning_message="WARNING: Java version 1.8 is recommended.
See https://docs.confluent.io/current/installation/versions-interoperability.html"

    # The first segment of the version number, which is '1' for releases before Java 9
    # it then becomes '9', '10', ...
    # Some examples of the first line of `java --version`:
    # 8 -> java version "1.8.0_152"
    # 9.0.4 -> java version "9.0.4"
    # 10 -> java version "10" 2018-03-20
    # 10.0.1 -> java version "10.0.1" 2018-04-17
    # Need to match the end of the line to prevent sed from printing the characters that do not match
    local java_version=$(${JAVA} -version 2>&1 | sed -E -n 's/.* version "([0-9]*).*$/\1/p')
    if [[ "${java_version}" -eq "1" ]]; then
        java_version=$(${JAVA} -version 2>&1 | sed -E -n 's/.* version "1.([0-9]*).*$/\1/p')
    fi

    if [[ "${java_version}" -lt "8" ]]; then
        die "${warning_message}"
    fi

    if [[ "${target_service}" = "zookeeper" || "${target_service}" = "kafka" ]]; then
        return
    fi

    if [[ "${java_version}" -ge "9" ]]; then
        cat <<EOF
Current Java version '${java_version}' is unsupported at this time. Confluent CLI will exit.

EOF
        die "${warning_message}"
    fi
}

start_command() {
    validate_java_version "${1}"
    validate_os_version "${1}"
    set_or_get_current
    echo "Using CONFLUENT_CURRENT: ${confluent_current}"
    start_or_stop_service "start" "${@}"
}

stop_command() {
    set_or_get_current
    echo "Using CONFLUENT_CURRENT: ${confluent_current}"
    start_or_stop_service "stop" "${@}"
}

start_or_stop_service() {
    local start_or_stop="${1}"

    local service="${2}"  # e.g. "kafka" or ""

    if [[ -n "${service}" ]]; then
        ! service_exists "${service}" && die "Unknown service: ${service}"
        local dep_var_name=`echo "deps_${start_or_stop}_${service}" | sed "s/-/_/"`
        local service_dep=( ${dep_var_name}[@] )
        local deps_list=( ${!service_dep} )
    else
        # Did not specify a service, so do them all
        is_confluent_platform
        status=$?
        if [[ "${start_or_stop}" = "start" ]]; then
            if [[ ${status} -eq 0 ]]; then
                local deps_list=( "${services[@]}" "${confluent_platform_services[@]}" )
            else
                local deps_list=( "${services[@]}" )
            fi
        else
            if [[ ${status} -eq 0 ]]; then
                local deps_list=( "${rev_confluent_platform_services}" "${rev_services[@]}" )
            else
                local deps_list=( "${rev_services[@]}" )
            fi
        fi
    fi

    local dependency=""
    shift
    for dependency in "${deps_list[@]}"; do
        "${start_or_stop}_${dependency}" "${@}";
    done
    if [[ -n "${service}" ]]; then
        "${start_or_stop}_${service}" "${@}";
    fi
}

status_command() {
    #TODO: consider whether a global call to this one with every invocation makes more sense
    set_or_get_current

    local command="${1}"

    service_exists "${command}"
    status=$?
    if [[ "x${command}" == "x" || ${status} -eq 0 ]]; then
        status_service "$@"
    elif [[ "${command}" == "connectors" ]]; then
        connect_subcommands "status"
    else
        connect_subcommands "status" "${@}"
    fi
}

print_current() {
    set_or_get_current
    echo "${confluent_current}"
}

check_demo_repo_uptodate() {
    local is_network_up="${1}"
    local repo="${2}"

    if [ $is_network_up == true ]; then
      get_version
      cd $repo && git fetch --tags && git checkout ${confluent_version}
    fi
}


is_demo_name_valid() {
    local demo_name="${1}"
    local repo="${2}"
    local subcommand="${3}"

    if [[ -z $demo_name ]]; then
      die "'confluent demo $subcommand' requires an argument that specifies the demo name. Please run 'confluent demo list' to see available demos"
    elif [[ ! -d $demo_name ]]; then
      die "Demo $demo_name does not exist in ${confluent_home}/$repo. Please run 'confluent demo list' to see available demos"
    fi

    return 0

}

demo_command() {
    local subcommand="${1}"
    local demo_name="${2}"
    local repo="examples"

    if [[ "x${subcommand}" == "x" || "${subcommand}" == "help" ]]; then
      demo_usage
    fi

    # Check network connectivity
    local is_network_up=false
    curl --silent --output /dev/null github.com
    status=$?
    if [[ ${status} -ne 0 ]]; then
      is_network_up=false
    else
      is_network_up=true
    fi

    # To populate ${confluent_version}
    get_version

    # Clone the repo
    if [[ ! -d "${confluent_home}/$repo" ]]; then
      if [ $is_network_up == true ]; then
        git clone https://github.com/confluentinc/$repo.git ${confluent_home}/$repo
        status=$?
        if [[ ${status} -ne 0 ]]; then
            die "'git clone https://github.com/confluentinc/$repo.git ${confluent_home}/$repo' failed. Please verify your git access"
        fi
        cd ${confluent_home}/$repo && git checkout ${confluent_version} 2> /dev/null
      else
        echo "'confluent demo $subcommand $demo_name' requires network connectivity. Please try again when you are connected."
        return
      fi
    fi

    cd ${confluent_home}/$repo
    status=$?
    if [[ ${status} -ne 0 ]]; then
        die "Cannot find ${confluent_home}/$repo"
        return
    fi

    if [[ $subcommand == "list" ]]; then
      check_demo_repo_uptodate $is_network_up "${confluent_home}/$repo"
      echo -e "Available demos from ${confluent_home}/$repo (start a demo 'confluent demo start <demo-name>':"
      ls | grep -v "README.md" | grep -v "utils" | grep -v "LICENSE"
    elif [[ $subcommand == "update" ]]; then
      if [ $is_network_up == true ]; then
        cd ${confluent_home}/$repo && git fetch --tags && git checkout ${confluent_version}
      else
        echo "Running 'confluent demo $subcommand $demo_name' requires network connectivity. Please try again when you are connected."
        return
      fi
    elif [[ $subcommand == "start" ]]; then
      check_demo_repo_uptodate $is_network_up "${confluent_home}/$repo"
      is_demo_name_valid "$demo_name" "$repo" "$subcommand"
      cd $demo_name
      ./start.sh
    elif [[ $subcommand == "stop" ]]; then
      check_demo_repo_uptodate $is_network_up "${confluent_home}/$repo"
      is_demo_name_valid "$demo_name" "$repo" "$subcommand"
      cd $demo_name
      ./stop.sh
    elif [[ $subcommand == "info" ]]; then
      check_demo_repo_uptodate $is_network_up "${confluent_home}/$repo"
      is_demo_name_valid "$demo_name" "$repo" "$subcommand"
      cat ${confluent_home}/${repo}/${demo_name}/README.md
    else
        invalid_argument "demo" "${subcommand}"
    fi

}

destroy_command() {
    if [[ -f "${confluent_current_dir}confluent.current" ]]; then
        export confluent_current="$( cat "${confluent_current_dir}confluent.current" )"
    fi

    [[ ${confluent_current} == ${confluent_current_dir}confluent* ]] \
        && stop_command \
        && echo "Deleting: ${confluent_current}" \
        && rm -rf "${confluent_current}" \
        && rm -f "${confluent_current_dir}confluent.current"
}

top_command() {
    set_or_get_current
    local service="${1}"

    if [[ -n "${service}" ]]; then
        ! service_exists "${service}" && die "Unknown service: ${service}"
    fi

    case "${platform}" in
        Darwin|Linux)
            top_"${platform}" "$@";;
        *)
            die "'top' not available in platform: ${platform}" "$@";;
    esac
}

top_Linux() {
    local service=( "${1}" )

    [[ -z "${service}" ]] && service=( "${services[@]}" )

    local pids=""
    local item=""
    for item in "${service[@]}"; do
        local service_dir="${confluent_current}/${item}"
        local service_pid="$( cat "${service_dir}/${item}.pid" 2> /dev/null )"
        [[ -n "${service_pid}" ]] && pids="${pids}${service_pid},"
    done
    pids="${pids%%','}"
    [[ -z "${pids}" ]] && die "No services are running"
    top -p "${pids}"
}

top_Darwin() {
    local service="${1}"

    [[ -z "${service}" ]] && die "Missing required service argument in '${command_name} top'"

    local service_dir="${confluent_current}/${service}"
    local service_pid="$( cat "${service_dir}/${service}.pid" 2> /dev/null )"
    top -pid "${service_pid}"
}

log_command() {
    set_or_get_current
    local service="${1}"

    [[ -z "${service}" ]] && die "Missing required service argument in '${command_name} log'"

    if [[ -n "${service}" ]]; then
        ! service_exists "${service}" && die "Unknown service: ${service}"
    fi
    shift

    local service_dir="${confluent_current}/${service}"
    local service_log="${service_dir}/${service}.stdout"

    if [[ $# -gt 0 ]]; then
        tail "${@}" "$service_log"
    else
        less "$service_log"
    fi
}

connect_bundled_command() {
    echo "Bundled Predefined Connectors (edit configuration under etc/):"

    local entry=""
    for entry in "${connector_properties[@]}"; do
        local key="${entry%%=*}"
        echo "  ${key}"
    done
}

connect_list_command() {
    local subcommand="${1}"

    is_running "connect" "false"
    status=$?
    if [[ ${status} -ne 0 ]]; then
        is_running "connect" "true"
    fi

    if [[ "${subcommand}" == "connectors" ]]; then
            connect_bundled_command
    elif [[ "${subcommand}" == "plugins" ]]; then
        echo "Available Connector Plugins: "
        curl --max-time "${_default_curl_timeout}" -s -X GET \
            http://localhost:"${connect_port}"/connector-plugins \
            | eval ${FORMAT_CMD}
    else
        invalid_argument "list" "${subcommand}"
    fi
}

connect_status_command() {
    local connector="${1}"

    is_running "connect" "false"
    status=$?
    if [[ ${status} -ne 0 ]]; then
        is_running "connect" "true" \
        || die "To get the status of connectors try starting 'connect' service first."
    fi

    if [[ -n "${connector}" ]]; then
        curl --max-time "${_default_curl_timeout}" -s -X GET \
            http://localhost:"${connect_port}"/connectors/"${connector}"/status \
            | eval ${FORMAT_CMD}
    else
        curl --max-time "${_default_curl_timeout}" -s -X GET \
            http://localhost:"${connect_port}"/connectors \
            | eval ${FORMAT_CMD}
    fi
}

connector_config_template() {
    local connector_name="${1}"
    local config_file="${2}"
    local nested="${3}"

    [[ ! -f "${config_file}" ]] \
        && die "Can't load connector configuration. Config file '${config_file}' does not exist."

    local config="{"
    # TODO: decide which name to use. The one in the file or the predefined
    #local name_line="$( grep ^name ${config_file} )"
    #name="${name_line#*=}"
    name="${connector}"

    append_key_value "name" "${name}"
    local wrapper=""
    [[ "${nested}" == true ]] && wrapper=" \"config\": {"

    config="${config}${_retval},${wrapper}"

    while IFS= read -r line; do
        local key="${line%%=*}"
        local value="${line#*=}"
        append_key_value "${key}" "${value}"
        config="${config}${_retval},"
    done < <( grep -v ^# "${config_file}" | grep -v ^name | grep -v -e '^[[:space:]]*$' | grep '=' )

    [[ "${nested}" == true ]] && wrapper="}"

    config="${config}}"
    config="${config%%',}'}"
    config="${config}}${wrapper}"
    _retval="$( echo "${config}" | eval ${FORMAT_CMD} )"
}

extract_name_from_properties_file() {
    local config_file="${1}"

    [[ ! -f "${config_file}" ]] \
        && die "Can't load connector configuration. Config file '${config_file}' does not exist."

    local name_line="$( grep ^name ${config_file} | grep '=' )"
    _retval="${name_line#*=}"
}

extract_bootstrapservers_from_properties_file() {
    local config_file="${1}"

    [[ ! -f "${config_file}" ]] \
        && die "Can't load Confluent Cloud configuration. Config file '${config_file}' does not exist."

    local bootstrap_servers="$( grep "^bootstrap.server" ${config_file} | grep '=' )"
    bootstrap_servers="${bootstrap_servers#*=}"
    # Remove backslash found in some Confluent Cloud configuration files
    bootstrap_servers=${bootstrap_servers/\\/}

    if [[ "x${bootstrap_servers}" == "x" ]]; then
      die "Can't read 'bootstrap.servers' from configuration file $config_file"
    fi

    _retval="${bootstrap_servers}"
}

extract_configfile_from_command() {

    # Default Confluent Cloud configuration file
    file="$HOME/.ccloud/config"

    regex="--config [^[:space:]]+"
    if [[ "${command}" =~ $regex ]]; then
      match="${BASH_REMATCH[0]}"
      file="${match#* }"
      # Remove config file from command
      command="${command/--config $file/}"
    fi
    regex="--config[[:space:]]*$"
    if [[ "${command}" =~ $regex ]]; then
      die "'--config' requires an argument that corresponds to the configuration file."
    fi

    _retval="${file}"
}

append_key_value() {
    local key="${1}"
    local value="${2}"

    _retval="\"${key}\": \"${value}\""
}

is_predefined_connector() {
    local connector_name="${1}"
    [[ -z "${connector_name}" ]] && die "Connector name is missing"

    _retval=""
    local entry=""
    for entry in "${connector_properties[@]}"; do
        local key="${entry%%=*}"
        local value="${entry#*=}"
        if [[ "${key}" == "${connector_name}" ]]; then
            _retval="${value}"
            return 0
        fi
    done
    return 1
}

connect_load_command() {
    local connector="${1}"
    local flag="${2}"
    local config_file="${3}"

    if [[ "x${connector}" == "x" ]]; then
        die "Missing required connector name argument in '${command_name} load'"
    elif [[ "x${flag}" == "x" ]]; then
        if is_predefined_connector "${connector}"; then
            connector_config_template "${connector}" "${confluent_conf}/${_retval}" "true"
            parsed_json="${_retval}"
        else
            die "\
${connector} is not a predefined connector name.
Use '${command_name} load ${connector} -d <connector-config-file.[json|properties]' to load the connector's configuration."
        fi
    else
        if [[ "${flag}" != "-d" ]]; then
            invalid_argument "load" "${flag}"
        fi

        [[ ! -f "${config_file}" ]] \
            && die "Can't load connector configuration. Config file '${config_file}' does not exist."

        # Check whether we have json contents.
        is_json "${config_file}"
        status=$?

        local parsed_json=""
        if [[ ${status} -eq 0 ]]; then
            # It's JSON format load it.
            extract_json_config "${config_file}" "false"
            parsed_json="${_retval}"
        else
            file "${config_file}" | grep "ASCII" > /dev/null 2>&1
            status=$?
            if [[ ${status} -eq 0 ]]; then
                # Potentially properties file. Try to load it.
                extract_name_from_properties_file "${config_file}"
                [[ "x${_retval}" == "x" ]] \
                    && die "Missing 'name' property from connectors properties file."
                local connector_name="${_retval}"
                connector_config_template "${connector_name}" "${config_file}" "true"
                parsed_json="${_retval}"
            else
                invalid_argument "config" "${config_file}"
            fi
        fi
    fi

    curl --max-time "${_default_curl_timeout}" -s -X POST -d "${parsed_json}" \
        --header "content-Type:application/json" \
        http://localhost:"${connect_port}"/connectors \
        | eval ${FORMAT_CMD}
}

connect_unload_command() {
    local connector="${1}"
    [[ -z "${connector}" ]] && die "Can't unload connector. Connector name is missing"

    curl --max-time "${_default_curl_timeout}" -s -X DELETE \
        http://localhost:"${connect_port}"/connectors/"${connector}"
}

extract_json_config() {
    local config_file="${1}"
    # If it is nested, extract only the config field.
    local only_config="${2}"

    local parsed_json=""
    # Treating here the JSON contents as flat json, or nested with a specific field called
    # "config" is good enough for now.
    if [[ ${only_config} == true ]]; then
        which jq > /dev/null 2>&1
        status=$?

        if [[ ${status} -ne 0 ]]; then
            die "Error: Parsing config from JSON file '${config_file}' failed."
        fi

        cat "${config_file}" | jq -e '.config' > /dev/null 2>&1
        status=$?

        if [[ ${status} -ne 0 ]]; then
            die "Error: Parsing JSON file '${config_file}' failed"
        fi

        parsed_json=$( cat "${config_file}" | jq -e '.config' )
    else
        parsed_json=$( cat "${config_file}" )
    fi
    _retval="${parsed_json}"
}

produce_command() {
    local topicname="${1}"
    shift
    local args="$@"

    if [[ "x${topicname}" == "x" ]]; then
        echo -e "Error: Missing required topic name argument in '${command_name} produce <topicname>'\n"
        produce_usage "produce" 1
    fi

    export command="${args}"
    local bootstrapserver=""
    if [[ ! "$command" =~ "broker-list" ]]; then
      export_kafka
      bootstrapserver="--broker-list localhost:${kafka_port}"
    fi
    cloud="--cloud"
    if [[ "$command" =~ "$cloud" ]]; then
      extract_configfile_from_command
      file="${_retval}"
      command=${command//$cloud/--producer.config $file}
      extract_bootstrapservers_from_properties_file "${file}"
      bootstrapserver="--broker-list ${_retval}"
    fi
    avro="--value-format avro"
    if [[ "$command" =~ "$avro" ]]; then
      command=${command//$avro/}
      LOG_DIR=${tmp_dir} ${confluent_home}/bin/kafka-avro-console-producer $bootstrapserver --topic $topicname $command
    else
      ${confluent_home}/bin/kafka-console-producer $bootstrapserver --topic $topicname $command
    fi
}

consume_command() {
    local topicname="${1}"
    shift
    local args="$@"

    if [[ "x${topicname}" == "x" ]]; then
        echo -e "Error: Missing required topic name argument in '${command_name} consume <topicname>'\n"
        consume_usage "consume" 1
    fi

    export command="${args}"
    local bootstrapserver=""
    if [[ ! "$command" =~ "bootstrap-server" ]]; then
      export_kafka
      bootstrapserver="--bootstrap-server localhost:${kafka_port}"
    fi
    cloud="--cloud"
    if [[ "$command" =~ "$cloud" ]]; then
      extract_configfile_from_command
      file="${_retval}"
      command=${command//$cloud/--consumer.config $file}
      extract_bootstrapservers_from_properties_file "${file}"
      bootstrapserver="--bootstrap-server ${_retval}"
    fi
    avro="--value-format avro"
    if [[ "$command" =~ "$avro" ]]; then
      command=${command//$avro/}
      LOG_DIR=${tmp_dir} SCHEMA_REGISTRY_LOG4J_LOGGERS="INFO, stdout" ${confluent_home}/bin/kafka-avro-console-consumer $bootstrapserver --topic $topicname $command
    else
      ${confluent_home}/bin/kafka-console-consumer $bootstrapserver --topic $topicname $command
    fi
}

connect_config_command() {
    local connector="${1}"
    local flag="${2}"
    local config_file="${3}"

    if [[ "x${connector}" == "x" ]]; then
        die "Missing required connector name argument in '${command_name} config'"
    elif [[ "x${flag}" == "x" ]]; then
        echo "Current configuration of '${connector}' connector:"
        curl --max-time "${_default_curl_timeout}" -s -X GET \
            http://localhost:"${connect_port}"/connectors/"${connector}"/config \
            | eval ${FORMAT_CMD}
        return $?
    fi

    if [[ "${flag}" != "-d" ]]; then
        invalid_argument "config" "${flag}"
    fi

    [[ ! -f "${config_file}" ]] \
        && die "Can't load connector configuration. Config file '${config_file}' does not exist."

    # Check whether we have json contents.
    is_json "${config_file}"
    status=$?

    local parsed_json=""
    if [[ ${status} -eq 0 ]]; then
        # It's JSON format load it.
        extract_json_config "${config_file}" "true"
        parsed_json="${_retval}"
    else
        file "${config_file}" | grep "ASCII" > /dev/null 2>&1
        status=$?
        if [[ ${status} -eq 0 ]]; then
            # Potentially properties file. Try to load it.
            extract_name_from_properties_file "${config_file}"
            [[ "x${_retval}" == "x" ]] \
                && die "Missing 'name' property from connectors properties file."
            local connector_name="${_retval}"
            connector_config_template "${connector_name}" "${config_file}" "false"
            parsed_json="${_retval}"
        else
            invalid_argument "config" "${config_file}"
        fi
    fi

    curl --max-time "${_default_curl_timeout}" -s -X PUT \
        -H "Content-Type: application/json" \
        -d "${parsed_json}" \
        http://localhost:"${connect_port}"/connectors/"${connector}"/config \
        | eval ${FORMAT_CMD}
}

connect_restart_command() {
    echo "Not implemented yet!"
}

acl_command() {
    local service="${1}"
    shift
    case "${service}" in
        schema-registry)
                    if [[ -f "${confluent_bin}/sr-acl-cli" ]] ; then
                        exec_cli "sr-acl-cli" "--config" "${confluent_conf}/schema-registry/schema-registry.properties" "$@"
                    else
                        echo "Please install Confluent Security Plugins to use acl schema-registry"
                    fi;;
        *) die "Missing required <service> argument. Type 'confluent help acl' to get a list of services supporting ACL";;
    esac
}

connect_subcommands() {
    set_or_get_current
    export_connect

    local subcommand="${1}"

    case "${subcommand}" in
        list)
            shift
            connect_list_command "$@";;

        load)
            shift
            connect_load_command "$@";;

        unload)
            shift
            connect_unload_command "$@";;

        status)
            shift
            connect_status_command "$@";;

        config)
            shift
            connect_config_command "$@";;

        restart)
            shift
            connect_restart_command "$@";;

        *) invalid_subcommand "connect" "$@";;
    esac
}

version_command() {
    set_or_get_current
    local service="${1}"

    if [[ -n "${service}" ]]; then
        ! service_exists "${service}" && die "Unknown service: ${service}"
        if [[ "x${service}" == "xkafka" ]]; then
            echo "${kafka_version}"
        elif [[ "x${service}" == "xzookeeper" ]]; then
            echo "${zookeeper_version}"
        else
            echo "${confluent_version}"
        fi
    else
        echo "${confluent_flavor}: ${confluent_version}"
    fi
}

produce_usage() {
    local command="${1}"
    local exit_status="${2}"
    if [[ -z "${exit_status}" ]]; then
      exit_status=0
    fi

    cat <<EOF
Usage: ${command_name} produce <topicname> [--value-format avro --property value.schema=<schema>] [--cloud] [--config <filename>] [other optional args]

Description:
    Produce to Kafka topic specified by <topicname>.

    By default, this command produces to the Kafka cluster on the localhost.
    To send to another Kafka cluster, set the '--broker-list' argument.

    Confluent Cloud:
    To send to Confluent Cloud, set the '--cloud' argument, which assumes topic is already created.
    By default, it uses Confluent Cloud configuration file at $HOME/.ccloud/config
    To change the configuration file, set '--config <filename>'

    By default, this command sends non-Avro data.
    To send Avro data, specify '--value-format avro --property value.schema=<schema>'

    After typing the command, enter each message on a new line. Press 'ctrl-c' to finish.

Examples:
    confluent produce mytopic1

    confluent produce mytopic1 --value-format avro --property value.schema='{"type":"record","name":"myrecord","fields":[{"name":"f1","type":"string"}]}'

    confluent produce mytopic1 --cloud --property parse.key=true --property key.separator=,

Optional Arguments:

EOF

    ${confluent_home}/bin/kafka-console-producer
    exit $exit_status
}

consume_usage() {
    local command="${1}"
    local exit_status="${2}"
    if [[ -z "${exit_status}" ]]; then
      exit_status=0
    fi

    cat <<EOF
Usage: ${command_name} consume <topicname> ] [--value-format avro] [--cloud] [--config <filename>] [other optional args]

Description:
    Consume from Kafka topic specified by <topicname>.

    By default, this command consumes from the Kafka cluster on the localhost.
    To consume from another Kafka cluster, set the '--bootstrap-server' argument.

    Confluent Cloud:
    To consume from Confluent Cloud, set the '--cloud' argument, which assumes topic is already created.
    By default, it uses Confluent Cloud configuration file at $HOME/.ccloud/config
    To change the configuration file, set '--config <filename>'

    By default, this command reads non-Avro data.
    To read Avro data, specify '--value-format avro'

Examples:
    confluent consume mytopic1 --from-beginning

    confluent consume mytopic1 --value-format avro --from-beginning

    confluent consume mytopic1 --cloud --property print.key=true --from-beginning

Optional Arguments:
EOF

    ${confluent_home}/bin/kafka-console-consumer
    exit $exit_status
}

list_usage() {
    cat <<EOF
Usage: ${command_name} list [ plugins | connectors ]

Description:
    List all the available services or plugins.

    Without arguments it prints the list of all the available services.

    Given 'plugins' as subcommand, prints all the connector-plugins which are
    discoverable in the current Confluent Platform deployment.

    Given 'connectors' as subcommand, prints a list of connector names that map to predefined
    connectors. Their configuration files can be found under 'etc/' directory in Confluent Platform.


Examples:
    confluent list
        Prints the available services.

    confluent list plugins
        Prints all the connector plugins (connector classes) discoverable by Connect runtime.

    confluent list connectors
        Prints a list of predefined connectors.

EOF
    exit 0
}

start_usage() {
    cat <<EOF
Usage: ${command_name} start [<service>]

Description:
    Start all services. If a specific <service> is given as an argument, it starts this service
    along with all of its dependencies.


Output:
    Prints status messages after starting each service to indicate successful startup or error.


Examples:
    confluent start
        Starts all available services.

    confluent start kafka
        Starts kafka and zookeeper as its dependency.

EOF
    exit 0
}

stop_usage() {
    cat <<EOF
Usage: ${command_name} stop [<service>]

Description:
    Stop all services. If a specific <service> is given as an argument it stops this service
    along with all of its dependencies.


Output:
    Prints status messages after stopping each service to indicate successful shutdown or error.


Examples:
    confluent stop
        Stops all available services.

    confluent stop kafka
        Stops kafka and zookeeper as its dependency.

EOF
    exit 0
}

status_usage() {
    cat <<EOF
Usage: ${command_name} status [ <service> | connectors | <connector-name> ]

Description:
    Return the status of services or connectors.

    Without arguments it prints the status of all the available services.

    If a specific <service> is given as an argument the status of the requested service is returned
    along with the status of its dependencies.

    Given 'connectors' as subcommand, prints a list with the connectors currently loaded in Connect.

    If a specific <connector-name> is given, then the status of the requested connector is returned.


Examples:
    confluent status
        Prints the status of the available services.

    confluent status kafka
        Prints the status of the 'kafka' service.

    confluent status connectors
        Prints a list with the loaded connectors at any given moment.

    confluent status file-source
        Prints the status of the connector with the given name.

EOF
    exit 0
}

current_usage() {
    cat <<EOF
Usage: ${command_name} current

Description:
    Return the filesystem path of the data and logs of the services managed by the current
    confluent run. If such a path does not exist, it will be created.


Output:
    The filesystem directory path to the current confluent run.


Examples:
    confluent current
        /tmp/confluent.SpBP4fQi

EOF
    exit 0
}

demo_usage() {
    cat <<EOF
Usage: ${command_name} demo [ list | update | start | stop | info ] [ <demo-name> ]

Description:
    Confluent Platform demos from https://github.com/confluentinc/examples:

    - Running 'confluent demo' requires internet connectivity and GitHub access.
    - Demos are stored locally in ${confluent_home}/examples.
    - The demos are for sandbox testing in a development environment. Never run these demos in production.

    Subcommand options:

      'list'                : lists names of available demos
      'update'              : pulls the latest demo code from GitHub
      'start <demo-name>'   : starts the demo specified as the argument
      'stop <demo-name>'    : stops the demo specified as the argument
      'info <demo-name>'    : see README for specific demo

Examples:
    confluent demo list
        Lists names of available demos provided in the GitHub repo

    confluent demo update
        Pulls the latest demo code from the GitHub repo

    confluent demo start wikipedia
        Starts a demo called wikipedia

    confluent demo stop wikipedia
        Stops a demo called wikipedia

    confluent demo info wikipedia
        Provide README for demo called wikipedia

EOF
    exit 0
}

destroy_usage() {
    cat <<EOF
Usage: ${command_name} destroy

Description:
    Delete an existing confluent run. Any running services are stopped. The data and the log
    files of all services are deleted.


Examples:
    confluent destroy
        Confirms that every service is stopped and finally prints the filesystem path that is deleted.

EOF
    exit 0
}

log_usage() {
    cat <<EOF
Usage: ${command_name} log <service> [optional arguments to tail]

Description:
    Read or tail the log of a service. If no arguments are given, a snapshot of the log is opened with
    a viewer ('less' command). If any arguments are given, 'tail' is used instead and the arguments
    are passed to the tail command.


Examples:
    confluent log connect
        Opens the connect log using 'less'.

    confluent log kafka -f
        Tails the kafka log and waits to print additional output until the log command is interrupted.

EOF
    exit 0
}

top_usage() {
    cat <<EOF
Usage: ${command_name} top [<service>]

Description:
    Track resource usage of a service.

EOF
    exit 0
}

load_usage() {
    cat <<EOF
Usage: ${command_name} load [<connector-name> [-d <connector-config-file>]]

Description:
    Load a bundled connector with a predefined name or custom connector with a given configuration.

EOF
    exit 0
}

unload_usage() {
    cat <<EOF
Usage: ${command_name} unload [<connector-name>]

Description:
    Unload a connector with the given <connector-name>.

EOF
    exit 0
}

config_usage() {
    cat <<EOF
Usage: ${command_name} config <connector-name> [ -d <connector-config-file> ]

Description:
    Get or set a connector's configuration properties.

    Given only the connector's name, prints the connector's configuration if such a connector is
    currently loaded in Connect.

    Additionally, given a filename with the option '-d', it configures the connector '<connector-name>'.
    The file needs to be in a valid JSON or java properties format and has to contain a correct
    configuration for a connector with the same name as the one given in the command-line.


Examples:
    confluent config s3-sink
        Prints the current configuration of the predefined connector with name 's3-sink'

    confluent config wikipedia-file-source
        Prints the current configuration of a custom connector with name 'wikipedia-file-source'

    confluent config wikipedia-file-source -d ./wikipedia-file-source.json
        Configures a connector named 'wikipedia-file-source' by passing its configuration properties in
        JSON format.

    confluent config wikipedia-file-source -d ./wikipedia-file-source.properties
        Configures a connector named 'wikipedia-file-source' by passing its configuration properties as
        java properties.

EOF
    exit 0
}

acl_usage() {
    if [[ -z "${2}" ]]; then
    cat <<EOF
Usage: ${command_name} acl <service> [<parameters>]

Description:
    Specify ACL to a service. Use help acl <service> to get further details about
    [<parameters>]. Currently schema-registry is the only supported service for acl.

EOF
        exit 0
    else
        acl_command "${2}" "--help"
    fi
}

version_usage() {
    cat <<EOF
Usage: ${command_name} version [<service>]

Description:
    Print the Confluent Platform flavor and version, or the individual version of a service.

Examples:
    confluent version
        Prints the flavor and version of Confluent platform.

    confluent version kafka
        Prints the version of a service included with Confluent platform, 'kafka' in this example.

EOF
    exit 0
}

usage() {
    cat <<EOF
${command_name}: Manage a local Confluent Platform development environment.

Usage: ${command_name} <command> [<subcommand>] [<parameters>]

These are the available commands:

    acl         Specify acl for a service.
    config      Configure a connector.
    consume     Consume data from topics
    current     Get the path of the data and logs of the services managed by the current confluent run.
                Override default setting with "CONFLUENT_CURRENT" environment variable.
    demo        Run demos provided in GitHub repo https://github.com/confluentinc/examples
    destroy     Delete the data and logs of the current confluent run.
    list        List available services.
    load        Load a connector.
    log         Read or tail the log of a service.
    produce     Produce data to topics
    start       Start all services or a specific service along with its dependencies
    status      Get the status of all services or the status of a specific service along with its dependencies.
    stop        Stop all services or a specific service along with the services depending on it.
    top         Track resource usage of a service.
    unload      Unload a connector.
    version     Print the Confluent Platform flavor and version or the individual version of a service.

'${command_name} help' lists available commands. See '${command_name} help <command>' to read about a
specific command.

EOF
    exit 0
}

invalid_argument() {
    local command="${1}"
    local argument="${2}"
    die "Invalid argument '${argument}' given to '${command}'."
}

invalid_subcommand() {
    local command="${1}"
    shift
    echo "Unknown subcommand '${command} ${1}'."
    die "Type '${command_name} help ${command}' for a list of available subcommands."
}

invalid_command() {
    echo "Unknown command '${1}'."
    die "Type '${command_name} help' for a list of available commands."
}

invalid_requirement() {
    echo -n "'${command_name}' requires '${1}'"
    if [[ "x${2}" == "x" ]]; then
        echo "."
    else
        echo " >= '${2}'."
    fi

    exit ${ERROR_CODE}
}

cat >&2 <<EOF
    The local commands are intended for a single-node development environment
    only, NOT for production usage. https://docs.confluent.io/current/cli/index.html

EOF

help() {
    # Since this function performs essential initializations, call it as early as possible.
    validate_and_export_dir_layout "skip"

    command_exists "${1}" && ( "${1}"_usage "$@" || invalid_command "${1}" ) || usage
}

main() {
    # Since this function performs essential initializations, call it as early as possible.
    validate_and_export_dir_layout $@

    # Parse command-line arguments
    [[ $# -lt 1 ]] && usage

    requirements

    command="${1}"
    shift
    case "${command}" in
        help)
            if [[ -n "${1}" ]]; then
                help $@
            else
                usage
            fi;;

        config)
            connect_subcommands "${command}" "$@";;

        connect)
            connect_subcommands "$@";;

        consume)
            consume_command "$@";;

        current)
            print_current;;

        demo)
            demo_command "$@";;

        destroy)
            destroy_command;;

        list)
            list_command "$@";;

        load)
            connect_subcommands "${command}" "$@";;

        log)
            log_command "$@";;

        produce)
            produce_command "$@";;

        start)
            start_command "$@";;

        status)
            status_command "$@";;

        stop)
            stop_command "$@";;

        top)
            top_command "$@";;

        unload)
            connect_subcommands "${command}" "$@";;

        acl)
            acl_command "$@";;

        version)
            version_command "$@";;

        *) invalid_command "${command}";;
    esac

    success=true

    trap shutdown EXIT
}


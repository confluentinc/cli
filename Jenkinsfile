def config = jobConfig {
    nodeLabel = 'docker-oraclejdk8'
    properties = [
        parameters([
            string(name: 'TEST_PATH', defaultValue: 'muckrake/tests/acl_cli_test.py muckrake/tests/password_protection_cli_test.py muckrake/tests/rbac_kafka_cli_test.py', description: 'Use this to specify a test or subset of tests to run.'),
            string(name: 'NUM_WORKERS', defaultValue: '15', description: 'Number of EC2 nodes to use when running the tests.'),
            string(name: 'INSTALL_TYPE', defaultValue: 'source', choices: ['distro', 'source', 'tarball'], description: 'Use tarball or source or distro'),
            string(name: 'RESOURCE_URL', defaultValue: '', description: 'If using tarball or distro [deb, rpm], specify S3 URL to download artifacts from'),
            string(name: 'PARALLEL', defaultValue:'true', description: 'Whether to execute the tests in parallel. If disabled, you should probably reduce NUM_WORKERS')
        ])
    ]
    realJobPrefixes = ['cli']
    timeoutHours = 16
}

def job = {
    if (config.isPrJob) {
        configureGitSSH("github/confluent_jenkins", "private_key")

        stage('Setup Go and Build CLI') {
            withVaultEnv([["docker_hub/jenkins", "user", "DOCKER_USERNAME"],
                ["docker_hub/jenkins", "password", "DOCKER_PASSWORD"],
                ["github/confluent_jenkins", "user", "GIT_USER"],
                ["github/confluent_jenkins", "access_token", "GIT_TOKEN"],
                ["artifactory/tools_jenkins", "user", "TOOLS_ARTIFACTORY_USER"],
                ["artifactory/tools_jenkins", "password", "TOOLS_ARTIFACTORY_PASSWORD"],
                ["sonatype/confluent", "user", "SONATYPE_OSSRH_USER"],
                ["sonatype/confluent", "password", "SONATYPE_OSSRH_PASSWORD"]]) {
                withEnv(["GIT_CREDENTIAL=${env.GIT_USER}:${env.GIT_TOKEN}", "GIT_USER=${env.GIT_USER}", "GIT_TOKEN=${env.GIT_TOKEN}"]) {
                    withVaultFile([["maven/jenkins_maven_global_settings", "settings_xml",
                        "/home/jenkins/.m2/settings.xml", "MAVEN_GLOBAL_SETTINGS_FILE"],
                        ["gradle/gradle_properties_maven", "gradle_properties_file",
                        "gradle.properties", "GRADLE_PROPERTIES_FILE"]]) {
                        sh '''
                            wget "https://golang.org/dl/go1.14.7.linux-amd64.tar.gz" --quiet --output-document go1.14.7.tar.gz
                            tar -C $(pwd) -xzf go1.14.7.tar.gz
                            export GOROOT=$(pwd)/go
                            export GOPATH=$(pwd)/go/path
                            export GOBIN=$(pwd)/go/bin
                            export modulePath=$(pwd)/go/src/github.com/confluentinc/cli
                            mkdir -p $GOPATH/bin
                            mkdir -p $GOROOT/bin
                            export PATH=$GOPATH/bin:$GOROOT/bin:$PATH
                            echo "machine github.com\n\tlogin $GIT_USER\n\tpassword $GIT_TOKEN" > ~/.netrc
                            make deps
                            make build-confluent
                        '''
                    }
                }
            }
        }

        stage('Clone muckrake') {
            withVaultEnv([["docker_hub/jenkins", "user", "DOCKER_USERNAME"],
                ["docker_hub/jenkins", "password", "DOCKER_PASSWORD"],
                ["github/confluent_jenkins", "user", "GIT_USER"],
                ["github/confluent_jenkins", "access_token", "GIT_TOKEN"],
                ["artifactory/tools_jenkins", "user", "TOOLS_ARTIFACTORY_USER"],
                ["artifactory/tools_jenkins", "password", "TOOLS_ARTIFACTORY_PASSWORD"],
                ["sonatype/confluent", "user", "SONATYPE_OSSRH_USER"],
                ["sonatype/confluent", "password", "SONATYPE_OSSRH_PASSWORD"]]) {
                withEnv(["GIT_CREDENTIAL=${env.GIT_USER}:${env.GIT_TOKEN}"]) {
                    withVaultFile([["maven/jenkins_maven_global_settings", "settings_xml",
                        "/home/jenkins/.m2/settings.xml", "MAVEN_GLOBAL_SETTINGS_FILE"],
                        ["gradle/gradle_properties_maven", "gradle_properties_file",
                        "gradle.properties", "GRADLE_PROPERTIES_FILE"]]) {
                        sh '''
                            git clone git@github.com:confluentinc/muckrake.git
                        '''
                    }
                }
            }
        }

        stage('Build & Test Ducker Image') {
            def pem_file = ''
            pem_file = setupSSHKey("vagrant/instance_pem", "pem_file", "${env.WORKSPACE}/vagrant-instance.pem")
            withVaultEnv([["docker_hub/jenkins", "user", "DOCKER_USERNAME"],
                ["docker_hub/jenkins", "password", "DOCKER_PASSWORD"],
                ["github/confluent_jenkins", "user", "GIT_USER"],
                ["github/confluent_jenkins", "access_token", "GIT_TOKEN"],
                ["artifactory/tools_jenkins", "user", "TOOLS_ARTIFACTORY_USER"],
                ["artifactory/tools_jenkins", "password", "TOOLS_ARTIFACTORY_PASSWORD"],
                ["sonatype/confluent", "user", "SONATYPE_OSSRH_USER"],
                ["sonatype/confluent", "password", "SONATYPE_OSSRH_PASSWORD"]]) {
                withEnv(["GIT_CREDENTIAL=${env.GIT_USER}:${env.GIT_TOKEN}",
                    "AWS_KEYPAIR_FILE=${pem_file}", "GIT_BRANCH=master"]) {
                    withVaultFile([["maven/jenkins_maven_global_settings", "settings_xml",
                        "/home/jenkins/.m2/settings.xml", "MAVEN_GLOBAL_SETTINGS_FILE"],
                        ["gradle/gradle_properties_maven", "gradle_properties_file",
                        "gradle.properties", "GRADLE_PROPERTIES_FILE"]]) {
                        sh '''
                            if [ -z "${TEST_PATH}" ]; then
                                export TEST_PATH="muckrake/tests/everything_runs_test.py"
                            fi
                            muckrake/ducker/resources/setup-gradle-properties.sh
                            muckrake/ducker/resources/setup-git-credential-store
                            export CHANGE_BRANCH=master
                            export HASH=$(git rev-parse --short=7 HEAD)
                            sed -i "s?\(confluent-cli-\(.*\)=\)\(.*\)?\1$(pwd)/dist/confluent/confluent_SNAPSHOT-${HASH}_linux_amd64\.tar\.gz?" muckrake/ducker/ducker
                            cat muckrake/ducker/ducker
                            cd muckrake/ducker; CHANGE_BRANCH=master ./vagrant-build-ducker.sh --pr true
                        '''
                    }
                }
            }
        }
    }
}


def post = {
    if (config.isPrJob) {
        stage("Cleanup") {
            def pem_file = ''
            pem_file = setupSSHKey("vagrant/instance_pem", "pem_file", "${env.WORKSPACE}/vagrant-instance.pem")
            withEnv(["AWS_KEYPAIR_FILE=${pem_file}"]) {
                sh '''#!/usr/bin/env bash
                    pwd
                    ls
                    cd muckrake
                    cd ducker
                    . ./resources/aws-iam.sh
                    vagrant destroy -f
                '''
            }
        }
    }
}

runJob config, job, post

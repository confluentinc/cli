def config = jobConfig {
    nodeLabel = 'docker-oraclejdk8'
    properties = [
        parameters([
            string(name: 'TEST_PATH', defaultValue: 'muckrake/tests/everything_runs_test.py muckrake/tests/kafka_rest_mini_test.py', description: 'Use this to specify a test or subset of tests to run.'),
            string(name: 'NUM_WORKERS', defaultValue: '15', description: 'Number of EC2 nodes to use when running the tests.'),
            string(name: 'INSTALL_TYPE', defaultValue: 'source', choices: ['distro', 'source', 'tarball'], description: 'Use tarball or source or distro'),
            string(name: 'RESOURCE_URL', defaultValue: '', description: 'If using tarball or distro [deb, rpm], specify S3 URL to download artifacts from'),
            string(name: 'PARALLEL', defaultValue:'true', description: 'Whether to execute the tests in parallel. If disabled, you should probably reduce NUM_WORKERS')
        ])
    ]
    realJobPrefixes = ['cli']
    timeoutHours = 16
}

def pre = {
    if (config.isPrJob) {
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
                            git clone https://github.com/confluentinc/muckrake.git
                        '''
                    }
                }
            }
        }
    }
}



def job = {
    if (config.isPrJob) {
        stage('Build & Test Ducker Image') {
            cd muckrake
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
                    "AWS_KEYPAIR_FILE=${pem_file}"]) {
                    withVaultFile([["maven/jenkins_maven_global_settings", "settings_xml",
                        "/home/jenkins/.m2/settings.xml", "MAVEN_GLOBAL_SETTINGS_FILE"],
                        ["gradle/gradle_properties_maven", "gradle_properties_file",
                        "gradle.properties", "GRADLE_PROPERTIES_FILE"]]) {
                        sh '''
                            if [ -z "${TEST_PATH}" ]; then
                                export TEST_PATH="muckrake/tests/everything_runs_test.py"
                            fi
                            ducker/resources/setup-gradle-properties.sh
                            ducker/resources/setup-git-credential-store
                            cd ducker; ./vagrant-build-ducker.sh --pr true
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
                    cd muckrake
                    cd ducker
                    . ./resources/aws-iam.sh
                    vagrant destroy -f
                '''
            }
        }
    }
}

runJob config, pre, job, post

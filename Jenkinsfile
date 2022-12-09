def config = jobConfig {
    nodeLabel = 'docker-debian-10-system-test-jdk8'
    properties = [
        parameters([
            string(name: 'TEST_PATH', defaultValue: 'muckrake/tests/cli_acls_test.py muckrake/tests/cli_brokers_test.py muckrake/tests/cli_partitions_test.py muckrake/tests/cli_topics_test.py muckrake/tests/password_protection_cli_test.py muckrake/tests/rbac_kafka_cli_test.py', description: 'Use this to specify a test or subset of tests to run.'),
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
        def mavenSettingsFile = "/home/jenkins/.m2/settings.xml"
        withMavenSettings("maven/jenkins_maven_global_settings", "settings", "MAVEN_GLOBAL_SETTINGS_FILE", mavenSettingsFile) {

            stage('Setup Go and Build CLI') {
                writeFile file:'extract-iam-credential.sh', text:libraryResource('scripts/extract-iam-credential.sh')
                withVaultEnv([["docker_hub/jenkins", "user", "DOCKER_USERNAME"],
                    ["docker_hub/jenkins", "password", "DOCKER_PASSWORD"],
                    ["github/confluent_jenkins", "user", "GIT_USER"],
                    ["github/confluent_jenkins", "access_token", "GIT_TOKEN"],
                    ["artifactory/tools_jenkins", "user", "TOOLS_ARTIFACTORY_USER"],
                    ["artifactory/tools_jenkins", "password", "TOOLS_ARTIFACTORY_PASSWORD"],
                    ["sonatype/confluent", "user", "SONATYPE_OSSRH_USER"],
                    ["sonatype/confluent", "password", "SONATYPE_OSSRH_PASSWORD"],
                    ["aws/prod_cli_team", "key_id", "AWS_ACCESS_KEY_ID"],
                    ["aws/prod_cli_team", "access_key", "AWS_SECRET_ACCESS_KEY"]]){
                    withEnv(["GIT_CREDENTIAL=${env.GIT_USER}:${env.GIT_TOKEN}", "GIT_USER=${env.GIT_USER}", "GIT_TOKEN=${env.GIT_TOKEN}"]) {
                        sh '''#!/usr/bin/env bash
                            ip addr
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
        }
    }
}

runJob config, job, post

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
                    ["aws/prod_cli_team", "key_id", "AWS_ACCESS_KEY_ID"],
                    ["aws/prod_cli_team", "access_key", "AWS_SECRET_ACCESS_KEY"]]){
                    withEnv(["GIT_CREDENTIAL=${env.GIT_USER}:${env.GIT_TOKEN}", "GIT_USER=${env.GIT_USER}", "GIT_TOKEN=${env.GIT_TOKEN}"]) {
                        withGradleFile(["gradle/gradle_properties_maven", "gradle_properties_file",
                            "gradle.properties", "GRADLE_PROPERTIES_FILE"]) {
                            loadNPMCredentials()
                            sh '''#!/usr/bin/env bash
                                export HASH=$(git rev-parse --short=7 HEAD)
                                wget "https://golang.org/dl/go1.20.linux-amd64.tar.gz" --quiet --output-document go1.20.tar.gz
                                tar -C $(pwd) -xzf go1.20.tar.gz
                                export GOROOT=$(pwd)/go
                                export GOPATH=$(pwd)/go/path
                                export GOBIN=$(pwd)/go/bin
                                export modulePath=$(pwd)/go/src/github.com/confluentinc/cli
                                mkdir -p $GOPATH/bin
                                mkdir -p $GOROOT/bin
                                export PATH=$GOPATH/bin:$GOROOT/bin:$PATH
                                echo "machine github.com\n\tlogin $GIT_USER\n\tpassword $GIT_TOKEN" > ~/.netrc
                                go install github.com/goreleaser/goreleaser@v1.15.2
                                make build || exit 1
                                cd dist
                                dir=confluent_SNAPSHOT-${HASH}_linux_amd64
                                mv confluent_linux_amd64_v1 $dir
                                tarball=$dir.tar.gz
                                tar -czf $tarball $dir
                                aws s3api put-object --bucket confluent.cloud --key confluent-cli-system-test-builds/$tarball --body $tarball
                                aws s3api put-object-acl --bucket confluent.cloud --key confluent-cli-system-test-builds/$tarball --acl public-read
                            '''
                        }
                    }
                }
            }

            stage('Clone muckrake') {
                withVaultEnv([["docker_hub/jenkins", "user", "DOCKER_USERNAME"],
                    ["docker_hub/jenkins", "password", "DOCKER_PASSWORD"],
                    ["github/confluent_jenkins", "user", "GIT_USER"],
                    ["github/confluent_jenkins", "access_token", "GIT_TOKEN"]]) {
                    withEnv(["GIT_CREDENTIAL=${env.GIT_USER}:${env.GIT_TOKEN}"]) {
                        withGradleFile(["gradle/gradle_properties_maven", "gradle_properties_file",
                            "gradle.properties", "GRADLE_PROPERTIES_FILE"]) {
                            sh '''#!/usr/bin/env bash
                                export HASH=$(git rev-parse --short=7 HEAD)
                                export confluent_s3="https://s3-us-west-2.amazonaws.com"
                                git clone git@github.com:confluentinc/muckrake.git
                                cd muckrake
                                git checkout v3
                                sed -i "s?\\(confluent-cli-\\(.*\\)=\\)\\(.*\\)?\\1${confluent_s3}/confluent.cloud/confluent-cli-system-test-builds/confluent_SNAPSHOT-${HASH}_linux_amd64\\.tar\\.gz\\"?" ducker/ducker
                                sed -i "s?get_cli .*?& ${confluent_s3}/confluent.cloud/confluent-cli-system-test-builds/confluent_SNAPSHOT-${HASH}_linux_amd64\\.tar\\.gz?g" vagrant/base-ubuntu.sh
                                sed -i "s?get_cli .*?& ${confluent_s3}/confluent.cloud/confluent-cli-system-test-builds/confluent_SNAPSHOT-${HASH}_linux_amd64\\.tar\\.gz?g" vagrant/base-redhat.sh
                                git checkout -b cli_system_test_$HASH
                                git commit -am "System test configuration for CLI build ${HASH}"
                                git push -u origin cli_system_test_$HASH
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
                    ["github/confluent_jenkins", "access_token", "GIT_TOKEN"]]) {
                    withEnv(["GIT_CREDENTIAL=${env.GIT_USER}:${env.GIT_TOKEN}",
                        "AWS_KEYPAIR_FILE=${pem_file}", "GIT_BRANCH=v3"]) {
                        withGradleFile(["gradle/gradle_properties_maven", "gradle_properties_file",
                            "gradle.properties", "GRADLE_PROPERTIES_FILE"]) {
                            withMavenSettings("maven/jenkins_maven_global_settings", "settings", "MAVEN_GLOBAL_SETTINGS_FILE", mavenSettingsFile) {
                                loadNPMCredentials()
                                sh '''#!/usr/bin/env bash
                                    export HASH=$(git rev-parse --short=7 HEAD)
                                    . extract-iam-credential.sh
                                    if [ -z "${TEST_PATH}" ]; then
                                        export TEST_PATH="muckrake/tests/everything_runs_test.py"
                                    fi
                                    muckrake/ducker/resources/setup-gradle-properties.sh
                                    muckrake/ducker/resources/setup-git-credential-store
                                    export CHANGE_BRANCH=cli_system_test_$HASH
                                    cd muckrake/ducker; ./vagrant-build-ducker.sh --pr true
                                '''
                            }
                        }
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
                    export HASH=$(git rev-parse --short=7 HEAD)
                    . extract-iam-credential.sh
                    cd muckrake
                    cd ducker
                    . ./resources/aws-iam.sh
                    vagrant destroy -f
                    cd ..
                    git push --delete origin cli_system_test_$HASH
                '''
            }
            withVaultEnv([["aws/prod_cli_team", "key_id", "AWS_ACCESS_KEY_ID"],
                         ["aws/prod_cli_team", "access_key", "AWS_SECRET_ACCESS_KEY"]]){
                sh '''#!/usr/bin/env bash
                    export HASH=$(git rev-parse --short=7 HEAD)
                    aws s3 rm s3://confluent.cloud/confluent-cli-system-test-builds/confluent_SNAPSHOT-${HASH}_linux_amd64.tar.gz
                '''
            }
        }
    }
}

runJob config, job, post

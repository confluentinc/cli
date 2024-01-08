# SDK generation steps

- The generatation of the the SDK code requires the latest openapi spec yaml file.
- Current openapi spec file for the sdk can be found in api/openapi.yaml
- The latest openapi spec file for schema-registry endpoints can be found in schema-registry repository of confluentinc here : https://github.com/confluentinc/schema-registry/blob/master/core/generated/swagger-ui/schema-registry-api-spec.yaml
- Clone the ath-add-interfaces branch of the confluentinc/openapi-generator repo.
```sh
git clone https://github.com/confluentinc/openapi-generator.git
cd openapi-generator
git checkout ath-add-interfaces
```
- In openapi-generator create the required jar using `mvn clean package`
- Run the following command to generate the latest SDK code:
```sh
java -jar modules/openapi-generator-cli/target/openapi-generator-cli.jar generate \
 -i <path-to-opnapi-spec.yaml>  \
 -g go \
 -o <path-to-schema-registry-sdk-go-directory> \
 --package-name schemaregistry --git-user-id confluentinc \
 --git-repo-id schema-registry-sdk-go \
 --additional-properties=generateInterfaces=true \
 --additional-properties=nullables=true
 ```
- Lastly, run `make mock` from schema-registry-sdk-go directory to generate the api_default.go file.
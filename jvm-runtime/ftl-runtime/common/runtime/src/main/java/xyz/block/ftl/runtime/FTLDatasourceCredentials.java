package xyz.block.ftl.runtime;

import java.util.Map;

import jakarta.inject.Named;
import jakarta.inject.Singleton;

import io.quarkus.credentials.CredentialsProvider;

@Named(FTLDatasourceCredentials.NAME)
@Singleton
public class FTLDatasourceCredentials implements CredentialsProvider {

    public static final String NAME = "ftl-datasource-credentials";

    @Override
    public Map<String, String> getCredentials(String credentialsProviderName) {
        FTLController.Datasource datasource = FTLController.instance().getDatasource(credentialsProviderName);
        if (datasource == null) {
            return null;
        }
        return Map.of("user", datasource.username(), "password", datasource.password());
    }
}

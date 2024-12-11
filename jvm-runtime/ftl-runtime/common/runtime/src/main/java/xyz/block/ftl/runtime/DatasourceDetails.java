package xyz.block.ftl.runtime;

import java.net.URI;
import java.net.URISyntaxException;
import java.util.regex.Pattern;

import xyz.block.ftl.deployment.v1.GetDeploymentContextResponse;

public record DatasourceDetails(String connectionString, String username, String password) {

    public static DatasourceDetails fromDSN(String dsn, GetDeploymentContextResponse.DbType type) {
        String prefix = type.equals(GetDeploymentContextResponse.DbType.DB_TYPE_MYSQL) ? "jdbc:mysql" : "jdbc:postgresql";
        try {
            URI uri = new URI(dsn);
            String username = "";
            String password = "";
            String userInfo = uri.getUserInfo();
            if (userInfo != null) {
                var split = userInfo.split(":");
                username = split[0];
                password = split[1];
                return new DatasourceDetails(
                        new URI(prefix, null, uri.getHost(), uri.getPort(), uri.getPath(), uri.getQuery(), null)
                                .toASCIIString(),
                        username, password);
            } else {
                //TODO: this is horrible, just quick hack for now
                var matcher = Pattern.compile("[&?]user=([^?&]*)").matcher(dsn);
                if (matcher.find()) {
                    username = matcher.group(1);
                    dsn = matcher.replaceAll("");
                }
                matcher = Pattern.compile("[&?]password=([^?&]*)").matcher(dsn);
                if (matcher.find()) {
                    password = matcher.group(1);
                    dsn = matcher.replaceAll("");
                }
                matcher = Pattern.compile("^([^:]+):([^:]+)@").matcher(dsn);
                if (matcher.find()) {
                    username = matcher.group(1);
                    password = matcher.group(2);
                    dsn = matcher.replaceAll("");
                }
                matcher = Pattern.compile("tcp\\(([^:)]+):([^:)]+)\\)").matcher(dsn);
                if (matcher.find()) {
                    // Mysql has a messed up syntax
                    dsn = matcher.replaceAll(matcher.group(1) + ":" + matcher.group(2));
                }
                dsn = dsn.replaceAll("postgresql://", "");
                dsn = dsn.replaceAll("postgres://", "");
                dsn = dsn.replaceAll("mysql://", "");
                dsn = prefix + "://" + dsn;
                return new DatasourceDetails(dsn, username, password);
            }
        } catch (URISyntaxException e) {
            throw new RuntimeException(e);
        }
    }

}

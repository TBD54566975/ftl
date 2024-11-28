package xyz.block.ftl.deployment;

import org.jboss.jandex.DotName;

import xyz.block.ftl.Config;
import xyz.block.ftl.Cron;
import xyz.block.ftl.Enum;
import xyz.block.ftl.EnumHolder;
import xyz.block.ftl.Export;
import xyz.block.ftl.GeneratedRef;
import xyz.block.ftl.LeaseClient;
import xyz.block.ftl.Secret;
import xyz.block.ftl.Subscription;
import xyz.block.ftl.Topic;
import xyz.block.ftl.TypeAlias;
import xyz.block.ftl.TypeAliasMapper;
import xyz.block.ftl.Verb;

public class FTLDotNames {

    private FTLDotNames() {

    }

    public static final DotName SECRET = DotName.createSimple(Secret.class);
    public static final DotName CONFIG = DotName.createSimple(Config.class);
    public static final DotName EXPORT = DotName.createSimple(Export.class);
    public static final DotName ENUM = DotName.createSimple(Enum.class);
    public static final DotName ENUM_HOLDER = DotName.createSimple(EnumHolder.class);
    public static final DotName VERB = DotName.createSimple(Verb.class);
    public static final DotName CRON = DotName.createSimple(Cron.class);
    public static final DotName TYPE_ALIAS_MAPPER = DotName.createSimple(TypeAliasMapper.class);
    public static final DotName TYPE_ALIAS = DotName.createSimple(TypeAlias.class);
    public static final DotName SUBSCRIPTION = DotName.createSimple(Subscription.class);
    public static final DotName LEASE_CLIENT = DotName.createSimple(LeaseClient.class);
    public static final DotName GENERATED_REF = DotName.createSimple(GeneratedRef.class);
    public static final DotName TOPIC = DotName.createSimple(Topic.class);
}

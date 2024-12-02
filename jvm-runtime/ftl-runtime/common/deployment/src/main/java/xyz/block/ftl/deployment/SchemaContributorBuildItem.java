package xyz.block.ftl.deployment;

import java.util.ArrayList;
import java.util.List;
import java.util.function.Consumer;

import io.quarkus.builder.item.MultiBuildItem;
import xyz.block.ftl.schema.v1.Decl;

/**
 * A build item that contributes information to the final Schema
 */
public final class SchemaContributorBuildItem extends MultiBuildItem {

    final Consumer<ModuleBuilder> schemaContributor;

    public SchemaContributorBuildItem(Consumer<ModuleBuilder> schemaContributor) {
        this.schemaContributor = schemaContributor;
    }

    public SchemaContributorBuildItem(List<Decl> decls) {
        var data = new ArrayList<>(decls);
        this.schemaContributor = new Consumer<ModuleBuilder>() {
            @Override
            public void accept(ModuleBuilder moduleBuilder) {
                for (var i : data) {
                    moduleBuilder.addDecls(i);
                }
            }
        };
    }

    public Consumer<ModuleBuilder> getSchemaContributor() {
        return schemaContributor;
    }
}

package xyz.block.ftl.deployment;

import java.io.IOException;

import org.jboss.jandex.MethodInfo;
import org.jboss.logging.Logger;
import org.objectweb.asm.ClassReader;
import org.objectweb.asm.Opcodes;
import org.objectweb.asm.tree.AbstractInsnNode;
import org.objectweb.asm.tree.ClassNode;
import org.objectweb.asm.tree.LineNumberNode;
import org.objectweb.asm.tree.MethodNode;

import io.quarkus.gizmo.DescriptorUtils;
import xyz.block.ftl.schema.v1.Position;

public class PositionUtils {

    private static final Logger LOG = Logger.getLogger(PositionUtils.class);

    public static Position forMethod(MethodInfo method) {
        return getLineNumber(method.declaringClass().name().toString(), method);
    }

    public static Position forClass(String className) {
        return getLineNumber(className, null);
    }

    static Position getLineNumber(String className, MethodInfo method) {
        Position.Builder builder = Position.newBuilder();

        var cl = Thread.currentThread().getContextClassLoader();
        var cls = cl.getResource(className.replace('.', '/') + ".class");
        if (cls == null) {
            return builder.build();
        }

        try (var in = cls.openStream()) {
            ClassReader reader = new ClassReader(in);
            ClassNode clNode = new ClassNode(Opcodes.ASM9);
            reader.accept(clNode, Opcodes.ASM9);
            if (clNode.sourceFile == null) {
                return builder.build();
            }
            builder.setFilename(clNode.sourceFile);
            if (method != null) {
                var descriptor = DescriptorUtils.methodSignatureToDescriptor(method.returnType().descriptor(),
                        method.parameters().stream().map(p -> p.type().descriptor()).toArray(String[]::new));
                for (MethodNode mNode : clNode.methods) {
                    if (mNode.name.equals(method.name()) && mNode.desc.equals(descriptor)) {
                        for (AbstractInsnNode inNode : mNode.instructions) {
                            if (inNode instanceof LineNumberNode) {
                                builder.setLine(((LineNumberNode) inNode).line);
                                return builder.build();
                            }
                        }
                    }
                }
            }
            return builder.build();
        } catch (IOException e) {
            LOG.errorf(e, "Failed to read class %s", className);
            return builder.build();
        }
    }
}

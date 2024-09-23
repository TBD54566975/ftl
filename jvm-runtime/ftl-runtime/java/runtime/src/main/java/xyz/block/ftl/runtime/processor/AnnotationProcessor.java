package xyz.block.ftl.runtime.processor;

import java.io.BufferedWriter;
import java.io.IOException;
import java.io.OutputStreamWriter;
import java.nio.charset.StandardCharsets;
import java.util.Base64;
import java.util.HashMap;
import java.util.Map;
import java.util.Optional;
import java.util.Set;
import java.util.regex.Pattern;
import java.util.stream.Collectors;

import javax.annotation.processing.Completion;
import javax.annotation.processing.ProcessingEnvironment;
import javax.annotation.processing.Processor;
import javax.annotation.processing.RoundEnvironment;
import javax.lang.model.SourceVersion;
import javax.lang.model.element.AnnotationMirror;
import javax.lang.model.element.Element;
import javax.lang.model.element.ElementKind;
import javax.lang.model.element.ExecutableElement;
import javax.lang.model.element.TypeElement;
import javax.lang.model.element.VariableElement;
import javax.tools.Diagnostic;
import javax.tools.FileObject;
import javax.tools.StandardLocation;

import xyz.block.ftl.Config;
import xyz.block.ftl.Export;
import xyz.block.ftl.Secret;
import xyz.block.ftl.Verb;

/**
 * POC annotation processor for capturing JavaDoc, this needs a lot more work.
 */
public class AnnotationProcessor implements Processor {
    private static final Pattern REMOVE_LEADING_SPACE = Pattern.compile("^ ", Pattern.MULTILINE);
    private ProcessingEnvironment processingEnv;

    final Map<String, String> saved = new HashMap<>();

    @Override
    public Set<String> getSupportedOptions() {
        return Set.of();
    }

    @Override
    public Set<String> getSupportedAnnotationTypes() {
        return Set.of(Verb.class.getName(), Export.class.getName());
    }

    @Override
    public SourceVersion getSupportedSourceVersion() {
        return SourceVersion.latestSupported();
    }

    @Override
    public void init(ProcessingEnvironment processingEnv) {
        this.processingEnv = processingEnv;
    }

    @Override
    public boolean process(Set<? extends TypeElement> annotations, RoundEnvironment roundEnv) {
        //TODO: @VerbName, HTTP, CRON etc
        roundEnv.getElementsAnnotatedWithAny(Set.of(Verb.class, Export.class))
                .forEach(element -> {
                    Optional<String> javadoc = getJavadoc(element);

                    javadoc.ifPresent(doc -> {
                        String key = element.getSimpleName().toString();

                        if (element.getKind() == ElementKind.METHOD) {
                            saved.put("verb." + key, doc);
                        } else if (element.getKind() == ElementKind.CLASS) {
                            saved.put("data." + key, doc);
                        } else if (element.getKind() == ElementKind.ENUM) {
                            saved.put("enum." + key, doc);
                        }

                        if (element.getKind() == ElementKind.METHOD) {
                            var executableElement = (ExecutableElement) element;
                            executableElement.getParameters().forEach(param -> {
                                Config config = param.getAnnotation(Config.class);
                                if (config != null) {
                                    saved.put("config." + config.value(), extractCommentForParam(doc, param));
                                }
                                Secret secret = param.getAnnotation(Secret.class);
                                if (secret != null) {
                                    saved.put("secret." + secret.value(), extractCommentForParam(doc, param));
                                }
                            });
                        }
                    });
                });

        if (roundEnv.processingOver()) {
            write("META-INF/ftl-verbs.txt", saved.entrySet().stream().map(
                    e -> e.getKey() + "=" + Base64.getEncoder().encodeToString(e.getValue().getBytes(StandardCharsets.UTF_8)))
                    .collect(Collectors.toSet()));
        }
        return false;
    }

    /**
     * This method uses the annotation processor Filer API and we shouldn't use a Path as paths containing \ are not supported.
     */
    public void write(String filePath, Set<String> set) {
        if (set.isEmpty()) {
            return;
        }
        try {
            final FileObject listResource = processingEnv.getFiler().createResource(StandardLocation.CLASS_OUTPUT, "",
                    filePath.toString());

            try (BufferedWriter writer = new BufferedWriter(
                    new OutputStreamWriter(listResource.openOutputStream(), StandardCharsets.UTF_8))) {
                for (String className : set) {
                    writer.write(className);
                    writer.newLine();
                }
            }
        } catch (IOException e) {
            processingEnv.getMessager().printMessage(Diagnostic.Kind.ERROR, "Failed to write " + filePath + ": " + e);
            return;
        }
    }

    @Override
    public Iterable<? extends Completion> getCompletions(Element element, AnnotationMirror annotation, ExecutableElement member,
            String userText) {
        return null;
    }

    public Optional<String> getJavadoc(Element e) {
        String docComment = processingEnv.getElementUtils().getDocComment(e);

        if (docComment == null || docComment.isBlank()) {
            return Optional.empty();
        }

        // javax.lang.model keeps the leading space after the "*" so we need to remove it.

        return Optional.of(REMOVE_LEADING_SPACE.matcher(docComment)
                .replaceAll("")
                .trim());
    }

    /**
     * Read the @param tag in a JavaDoc comment to extract Config and Secret comments
     */
    private String extractCommentForParam(String doc, VariableElement param) {
        String variableName = param.getSimpleName().toString();
        int startIdx = doc.indexOf("@param " + variableName + " ");
        if (startIdx != -1) {
            int endIndex = doc.indexOf("\n", startIdx);
            if (endIndex == -1) {
                endIndex = doc.length();
            }
            return doc.substring(startIdx + variableName.length() + 8, endIndex);
        }
        return null;
    }

}

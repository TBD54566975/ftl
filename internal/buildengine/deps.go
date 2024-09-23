package buildengine

// TODO: remove this file

// func extractDependencies(module Module) ([]string, error) {
// 	switch module.Config.Language {
// 	case "go":
// 		// return extractGoFTLImports(module.Config.Module, module.Config.Dir)

// 	case "java", "kotlin":
// 		return extractJVMFTLImports(module.Config.Module, module.Config.Dir)

// 	case "rust":
// 		return extractRustFTLImports(module.Config.Module, module.Config.Dir)

// 	default:
// 		return nil, fmt.Errorf("unsupported language: %s", module.Config.Language)
// 	}
// }

// func extractKotlinFTLImports(self, dir string) ([]string, error) {
// 	dependencies := map[string]bool{}
// 	kotlinImportRegex := regexp.MustCompile(`^import ftl\.([A-Za-z0-9_.]+)`)

// 	err := filepath.WalkDir(filepath.Join(dir, "src/main/kotlin"), func(path string, d fs.DirEntry, err error) error {
// 		if err != nil {
// 			return err
// 		}
// 		if d.IsDir() || !(strings.HasSuffix(path, ".kt") || strings.HasSuffix(path, ".kts")) {
// 			return nil
// 		}
// 		file, err := os.Open(path)
// 		if err != nil {
// 			return err
// 		}
// 		defer file.Close()

// 		scanner := bufio.NewScanner(file)
// 		for scanner.Scan() {
// 			matches := kotlinImportRegex.FindStringSubmatch(scanner.Text())
// 			if len(matches) > 1 {
// 				module := strings.Split(matches[1], ".")[0]
// 				if module == self {
// 					continue
// 				}
// 				dependencies[module] = true
// 			}
// 		}
// 		return scanner.Err()
// 	})

// 	if err != nil {
// 		return nil, fmt.Errorf("%s: failed to extract dependencies from Kotlin module: %w", self, err)
// 	}
// 	modules := maps.Keys(dependencies)
// 	sort.Strings(modules)
// 	return modules, nil
// }

// func extractJVMFTLImports(self, dir string) ([]string, error) {
// 	dependencies := map[string]bool{}
// 	// We also attempt to look at kotlin files
// 	// As the Java module supports both
// 	kotin, kotlinErr := extractKotlinFTLImports(self, dir)
// 	if kotlinErr == nil {
// 		// We don't really care about the error case, its probably a Java project
// 		for _, imp := range kotin {
// 			dependencies[imp] = true
// 		}
// 	}
// 	javaImportRegex := regexp.MustCompile(`^import ftl\.([A-Za-z0-9_.]+)`)

// 	err := filepath.WalkDir(filepath.Join(dir, "src/main/java"), func(path string, d fs.DirEntry, err error) error {
// 		if err != nil {
// 			return fmt.Errorf("failed to walk directory: %w", err)
// 		}
// 		if d.IsDir() || !(strings.HasSuffix(path, ".java")) {
// 			return nil
// 		}
// 		file, err := os.Open(path)
// 		if err != nil {
// 			return fmt.Errorf("failed to open file: %w", err)
// 		}
// 		defer file.Close()

// 		scanner := bufio.NewScanner(file)
// 		for scanner.Scan() {
// 			matches := javaImportRegex.FindStringSubmatch(scanner.Text())
// 			if len(matches) > 1 {
// 				module := strings.Split(matches[1], ".")[0]
// 				if module == self {
// 					continue
// 				}
// 				dependencies[module] = true
// 			}
// 		}
// 		return scanner.Err()
// 	})

// 	// We only error out if they both failed
// 	if err != nil && kotlinErr != nil {
// 		return nil, fmt.Errorf("%s: failed to extract dependencies from Java module: %w", self, err)
// 	}
// 	modules := maps.Keys(dependencies)
// 	sort.Strings(modules)
// 	return modules, nil
// }

// func extractRustFTLImports(self, dir string) ([]string, error) {
// 	fmt.Fprintf(os.Stderr, "RUST TODO extractRustFTLImports\n")

// 	return nil, nil
// }

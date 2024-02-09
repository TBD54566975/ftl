const { execSync } = require("child_process");

try {
  console.log("Resolving Maven dependencies");
  execSync("mvn dependency:resolve --batch-mode", { stdio: "inherit" });
  console.log("Resolving Go dependencies");
  execSync("go mod download -x", { stdio: "inherit" });
  console.log("Dependency resolution successful");
} catch (error) {
  console.error("Dependency resolution failed:", error.message);
  process.exit(1);
}

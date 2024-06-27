// swift-tools-version: 5.10
// The swift-tools-version declares the minimum version of Swift required to build this package.

import PackageDescription

let package = Package(
    name: "Main",
    platforms: [
              .macOS(.v13),
          ],
    dependencies: [
        .package(url: "https://github.com/apple/swift-argument-parser.git", from: "1.2.0"),
        .package(path: "../../.."),
        .package(path: "../../../../../../swift-runtime/FTL"),
        .package(url:"https://github.com/grpc/grpc-swift.git", from: "1.23.0")
    ],
    targets: [
        // Targets are the basic building blocks of a package, defining a module or a test suite.
        // Targets can depend on other targets in this package and products from dependencies.
        .executableTarget(
            name: "main",
            dependencies: [
                .product(name: "ArgumentParser", package: "swift-argument-parser"),
                .product(name: "Stime", package: "Stime"),
                .product(name: "GRPC", package: "grpc-swift"),
                .product(name: "FTL", package: "FTL"),
            ]
        ),
    ]
)

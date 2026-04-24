use std::env;
use std::path::PathBuf;

fn main() {
    let manifest_dir = env::var("CARGO_MANIFEST_DIR").unwrap();
    let target_os = env::var("CARGO_CFG_TARGET_OS").unwrap();
    let target_arch = env::var("CARGO_CFG_TARGET_ARCH").unwrap();

    // Map to our dist folder naming
    let go_os = match target_os.as_str() {
        "macos" => "darwin",
        "windows" => "windows",
        _ => "linux",
    };
    
    let go_arch = match target_arch.as_str() {
        "x86_64" => "amd64",
        "aarch64" => "arm64",
        _ => &target_arch,
    };

    let lib_dir = PathBuf::from(manifest_dir)
        .join("dist")
        .join(format!("{}-{}", go_os, go_arch));

    // Tell cargo where to find the library
    println!("cargo:rustc-link-search=native={}", lib_dir.display());
    println!("cargo:rustc-link-lib=dylib=vsmprotocol");

    // Re-run if we change the binaries
    println!("cargo:rerun-if-changed=dist/");
}

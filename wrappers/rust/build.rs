use std::env;
use std::path::PathBuf;
use std::process::Command;

fn main() {
    let manifest_dir = env::var("CARGO_MANIFEST_DIR").unwrap();
    let target_os = env::var("CARGO_CFG_TARGET_OS").unwrap();
    let target_arch = env::var("CARGO_CFG_TARGET_ARCH").unwrap();
    
    let version = "0.1.5"; // Match the GitHub Release tag

    let (go_os, ext) = match target_os.as_str() {
        "macos" => ("darwin", ".dylib"),
        "windows" => ("windows", ".dll"),
        _ => ("linux", ".so"),
    };
    
    let go_arch = match target_arch.as_str() {
        "x86_64" => "amd64",
        "aarch64" => "arm64",
        _ => &target_arch,
    };

    let lib_basename = format!("vsmprotocol-{}-{}", go_os, go_arch);
    let lib_filename = format!("{}-{}", lib_basename, ext); // Original binary name
    let asset_filename = format!("{}.{}", lib_basename, ext.trim_start_matches('.'));
    
    let out_dir = PathBuf::from(env::var("OUT_DIR").unwrap());
    let dest_path = out_dir.join(format!("vsmprotocol{}", ext));

    if !dest_path.exists() {
        let url = format!(
            "https://github.com/cezarpena/vsm-protocol/releases/download/v{}/{}",
            version, asset_filename
        );
        
        println!("cargo:warning=Downloading VSM binary from {}", url);
        
        // Use curl (built-in on macOS/Linux/W10+) to download
        let status = Command::new("curl")
            .args(["-L", "-o", dest_path.to_str().unwrap(), &url])
            .status()
            .expect("Failed to download VSM binary with curl");

        if !status.success() {
            panic!("Failed to download VSM binary from GitHub Release v{}", version);
        }
    }

    println!("cargo:rustc-link-search=native={}", out_dir.display());
    println!("cargo:rustc-link-lib=dylib=vsmprotocol");
}

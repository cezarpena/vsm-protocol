//! VSM Stealth Protocol - Rust Wrapper
//!
//! Uses raw FFI to call the Go-compiled shared library.
//!
//! Add to Cargo.toml:
//! ```toml
//! [dependencies]
//! libloading = "0.8"
//! ```

use std::ffi::{c_char, c_int, c_longlong, c_uchar, CStr, CString};
use std::fmt;

use libloading::{Library, Symbol};

// C callback signatures
type KnockCallbackFn = extern "C" fn(c_int, *const c_char);
type MessageCallbackFn = extern "C" fn(c_int, *const c_char, *const c_char);

/// A loaded instance of the VSM Protocol library.
pub struct VSMProtocol {
    _lib: Library,
    generate_identity: Symbol<'static, unsafe extern "C" fn() -> *mut c_char>,
    free_string: Symbol<'static, unsafe extern "C" fn(*mut c_char)>,
    start_server: Symbol<
        'static,
        unsafe extern "C" fn(c_longlong, *const c_char, KnockCallbackFn, MessageCallbackFn),
    >,
    send_message: Symbol<'static, unsafe extern "C" fn(c_longlong, *const c_char) -> c_uchar>,
    dial: Symbol<
        'static,
        unsafe extern "C" fn(*const c_char, *const c_char, c_longlong, *const c_char) -> c_longlong,
    >,
}

#[derive(Debug)]
pub enum VSMError {
    LibraryLoad(String),
    SymbolLoad(String),
    IdentityGeneration,
}

impl fmt::Display for VSMError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            VSMError::LibraryLoad(e) => write!(f, "Failed to load library: {}", e),
            VSMError::SymbolLoad(e) => write!(f, "Failed to load symbol: {}", e),
            VSMError::IdentityGeneration => write!(f, "Failed to generate identity"),
        }
    }
}

impl std::error::Error for VSMError {}

impl VSMProtocol {
    /// Load the VSM Protocol shared library from the given path.
    pub fn load(path: &str) -> Result<Self, VSMError> {
        unsafe {
            let lib = Library::new(path)
                .map_err(|e| VSMError::LibraryLoad(e.to_string()))?;

            // Transmute lifetimes so symbols live as long as the Library.
            // Safe because we keep _lib alive for the struct's lifetime.
            let generate_identity = std::mem::transmute(
                lib.get::<unsafe extern "C" fn() -> *mut c_char>(b"GenerateVSMIdentity")
                    .map_err(|e| VSMError::SymbolLoad(e.to_string()))?,
            );
            let free_string = std::mem::transmute(
                lib.get::<unsafe extern "C" fn(*mut c_char)>(b"FreeString")
                    .map_err(|e| VSMError::SymbolLoad(e.to_string()))?,
            );
            let start_server = std::mem::transmute(
                lib.get::<unsafe extern "C" fn(c_longlong, *const c_char, KnockCallbackFn, MessageCallbackFn)>(
                    b"StartVSMServer",
                )
                .map_err(|e| VSMError::SymbolLoad(e.to_string()))?,
            );
            let send_message = std::mem::transmute(
                lib.get::<unsafe extern "C" fn(c_longlong, *const c_char) -> c_uchar>(b"SendMessage")
                    .map_err(|e| VSMError::SymbolLoad(e.to_string()))?,
            );
            let dial = std::mem::transmute(
                lib.get::<unsafe extern "C" fn(*const c_char, *const c_char, c_longlong, *const c_char) -> c_longlong>(
                    b"VSMDial",
                )
                .map_err(|e| VSMError::SymbolLoad(e.to_string()))?,
            );

            Ok(Self {
                _lib: lib,
                generate_identity,
                free_string,
                start_server,
                send_message,
                dial,
            })
        }
    }

    /// Generate a new identity. Returns the JSON string.
    pub fn generate_identity(&self) -> Result<String, VSMError> {
        unsafe {
            let ptr = (self.generate_identity)();
            if ptr.is_null() {
                return Err(VSMError::IdentityGeneration);
            }
            let json = CStr::from_ptr(ptr).to_string_lossy().into_owned();
            (self.free_string)(ptr);
            Ok(json)
        }
    }

    /// Start a stealth server on the given port.
    pub fn start_server(
        &self,
        port: i32,
        identity_json: &str,
        on_knock: KnockCallbackFn,
        on_message: MessageCallbackFn,
    ) {
        let id = CString::new(identity_json).unwrap();
        unsafe {
            (self.start_server)(port as c_longlong, id.as_ptr(), on_knock, on_message);
        }
    }

    /// Dial a remote VSM server. Returns session ID or -1 on failure.
    pub fn dial(&self, device: &str, target_ip: &str, port: i32, identity_json: &str) -> i32 {
        let dev = CString::new(device).unwrap();
        let ip = CString::new(target_ip).unwrap();
        let id = CString::new(identity_json).unwrap();
        unsafe {
            (self.dial)(dev.as_ptr(), ip.as_ptr(), port as c_longlong, id.as_ptr()) as i32
        }
    }

    /// Send an encrypted message over an active session.
    pub fn send_message(&self, session_id: i32, text: &str) -> bool {
        let msg = CString::new(text).unwrap();
        unsafe { (self.send_message)(session_id as c_longlong, msg.as_ptr()) != 0 }
    }
}

// --- Callback examples ---
extern "C" fn on_knock(session_id: c_int, peer_id: *const c_char) {
    let peer = unsafe { CStr::from_ptr(peer_id) }.to_string_lossy();
    println!(" [RUST] Knock from {peer} (session {session_id})");
}

extern "C" fn on_message(session_id: c_int, peer_id: *const c_char, msg: *const c_char) {
    let peer = unsafe { CStr::from_ptr(peer_id) }.to_string_lossy();
    let text = unsafe { CStr::from_ptr(msg) }.to_string_lossy();
    println!(" [RUST] Message from {peer}: {text}");
}

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let vsm = VSMProtocol::load("./vsmprotocol.dylib")?;

    let identity = vsm.generate_identity()?;
    println!(" [RUST] Identity: {identity}");

    // Server example:
    // vsm.start_server(9999, &identity, on_knock, on_message);

    // Client example:
    // let sid = vsm.dial("lo0", "127.0.0.1", 9999, &identity);
    // if sid != -1 { vsm.send_message(sid, "Hello from Rust!"); }

    println!(" [RUST] VSM Protocol loaded successfully.");
    Ok(())
}

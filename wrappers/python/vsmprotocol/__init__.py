import ctypes
import os
import platform
import json
import time

def get_lib_path():
    system = platform.system().lower()  # darwin, linux, windows
    machine = platform.machine().lower()

    # Map platform names to Go conventions
    arch_map = {"x86_64": "amd64", "amd64": "amd64", "aarch64": "arm64", "arm64": "arm64"}
    arch = arch_map.get(machine, machine)

    if system == "darwin":
        lib_name = "vsmprotocol.dylib"
    elif system == "windows":
        lib_name = "vsmprotocol.dll"
    else:
        lib_name = "vsmprotocol.so"

    base_path = os.path.dirname(__file__)
    # Path when distributed as a package: vsmprotocol/dist/...
    return os.path.abspath(os.path.join(base_path, "dist", f"{system}-{arch}", lib_name))

# Load the library
_lib_path = get_lib_path()
_lib = ctypes.CDLL(_lib_path)

# 1. Define Callback Types
# argtypes: (session_id, peer_id)
KNOCK_CALLBACK = ctypes.CFUNCTYPE(None, ctypes.c_int, ctypes.c_char_p)
# argtypes: (session_id, peer_id, message)
MESSAGE_CALLBACK = ctypes.CFUNCTYPE(None, ctypes.c_int, ctypes.c_char_p, ctypes.c_char_p)

# 2. C API Definitions
_lib.GenerateVSMIdentity.restype = ctypes.c_void_p
_lib.FreeString.argtypes = [ctypes.c_void_p]

_lib.StartVSMServer.argtypes = [ctypes.c_int, ctypes.c_char_p, KNOCK_CALLBACK, MESSAGE_CALLBACK]

_lib.SendMessage.argtypes = [ctypes.c_int, ctypes.c_char_p]
_lib.SendMessage.restype = ctypes.c_bool

_lib.VSMDial.argtypes = [ctypes.c_char_p, ctypes.c_char_p, ctypes.c_int, ctypes.c_char_p]
_lib.VSMDial.restype = ctypes.c_int

# 3. High-Level Python API

def generate_identity():
    ptr = _lib.GenerateVSMIdentity()
    if not ptr: return None
    
    json_str = ctypes.cast(ptr, ctypes.c_char_p).value.decode("utf-8")
    _lib.FreeString(ptr)
    return json.loads(json_str)

def start_server(port, identity_json, on_knock, on_message):
    """
    Starts a background thread in Go that listens for stealth knocls.
    identity_json: Full identity dictionary from generate_identity()
    """
    # Keep references to callbacks to prevent garbage collection
    _k_ref = KNOCK_CALLBACK(on_knock)
    _m_ref = MESSAGE_CALLBACK(on_message)
    
    id_str = json.dumps(identity_json).encode('utf-8')
    _lib.StartVSMServer(port, id_str, _k_ref, _m_ref)
    
    return _k_ref, _m_ref

def dial(device, target_ip, target_port, my_identity_json):
    """
    Dials a VSM server.
    Returns: A Session ID (integer) on success, or -1 on failure.
    """
    # Just like start_server, we pass our identity as a JSON string
    id_str = json.dumps(my_identity_json).encode('utf-8')
    
    session_id = _lib.VSMDial(
        device.encode('utf-8'), 
        target_ip.encode('utf-8'), 
        target_port, 
        id_str
    )
    
    return session_id

def send_message(session_id, text):
    """Sends an encrypted message over an active VSM session."""
    return _lib.SendMessage(session_id, text.encode('utf-8'))

# 4. Demo Execution
if __name__ == "__main__":
    def my_knock_handler(session_id, peer_id):
        pid = peer_id.decode('utf-8')
        print(f"\n [PYTHON] !!! NEW PEER CONNECTED !!!")
        print(f" [PYTHON] Session ID: {session_id}")
        print(f" [PYTHON] Peer Identity: {pid}")
        
        # Test sending a reply immediately
        print(f" [PYTHON] Sending automated reply...")
        send_message(session_id, "Welcome to the Invisible Network.")

    def my_message_handler(session_id, peer_id, message):
        pid = peer_id.decode('utf-8')
        msg = message.decode('utf-8')
        print(f" [PYTHON] MESSAGE RECEIVED from {pid}: {msg}")

    # Generate a fresh identity
    print(" [PYTHON] Generating Server Identity...")
    id_data = generate_identity()
    print(f" [PYTHON] Server ID: {id_data['id']}")

    print(f" [PYTHON] Starting VSM Server on port 9999...")
    refs = start_server(9999, id_data, my_knock_handler, my_message_handler)

    try:
        print(" [PYTHON] Server running. Waiting for ghost traffic... (Ctrl+C to stop)")
        while True:
            time.sleep(1)
    except KeyboardInterrupt:
        print("\n [PYTHON] Stopping...")
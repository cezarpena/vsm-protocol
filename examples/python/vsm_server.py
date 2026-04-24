import sys
import os

# Allow importing from the wrappers directory
sys.path.append(os.path.abspath(os.path.join(os.path.dirname(__file__), '../../wrappers/python')))

import __init__ as vsm
import time
def on_knock(sid, peer): print(f" [SRV] Knock from {peer.decode()}")
def on_msg(sid, peer, msg): print(f" [SRV] Message: {msg.decode()}")
server_id_json = {
    "id": "7hKot6Oqrv0BkXA_K20uOTPzAfTNAJ6Cn5Qa6IKR5NA", 
    "privKey": "Srnp6yXcpeoYY7glMVgfXVuFtRZmUu_zh4fYbB5qrdvuEqi3o6qu_QGRcD8rbS45M_MB9M0AnoKflBrogpHk0A", # Paste the full JSON you got from the server
    "pubKey": "7hKot6Oqrv0BkXA_K20uOTPzAfTNAJ6Cn5Qa6IKR5NA"
}
# IMPORTANT: Keep the return value in a variable! 
# If you don't, Python will garbage collect the callbacks and the server will crash.
refs = vsm.start_server(9999, server_id_json, on_knock, on_msg)

print(" [SRV] Server running. Waiting for ghost traffic... (Ctrl+C to stop)")
while True: time.sleep(1)
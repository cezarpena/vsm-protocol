import vsmprotocol as vsm
import time
def on_knock(sid, peer): print(f" [SRV] Knock from {peer.decode()}")
def on_msg(sid, peer, msg): print(f" [SRV] Message: {msg.decode()}")
server_id_json = {
    "id": "7hKot6Oqrv0BkXA_K20uOTPzAfTNAJ6Cn5Qa6IKR5NA", 
    "privKey": "Srnp6yXcpeoYY7glMVgfXVuFtRZmUu_zh4fYbB5qrdvuEqi3o6qu_QGRcD8rbS45M_MB9M0AnoKflBrogpHk0A", # Paste the full JSON you got from the server
    "pubKey": "7hKot6Oqrv0BkXA_K20uOTPzAfTNAJ6Cn5Qa6IKR5NA"
}
vsm.start_server(9999, server_id_json, on_knock, on_msg)
while True: time.sleep(1)
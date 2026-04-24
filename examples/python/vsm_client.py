import vsmprotocol as vsm

server_id_json = {
    "id": "7hKot6Oqrv0BkXA_K20uOTPzAfTNAJ6Cn5Qa6IKR5NA", 
    "privKey": "Srnp6yXcpeoYY7glMVgfXVuFtRZmUu_zh4fYbB5qrdvuEqi3o6qu_QGRcD8rbS45M_MB9M0AnoKflBrogpHk0A", # Paste the full JSON you got from the server
    "pubKey": "7hKot6Oqrv0BkXA_K20uOTPzAfTNAJ6Cn5Qa6IKR5NA"
}

# Dial the local server
sid = vsm.dial("lo0", "127.0.0.1", 9999, server_id_json)
if sid != -1:
    print(f" [CLI] Connected! Session: {sid}")
    vsm.send_message(sid, "Hello from the Invisible Client")
else:
    print(" [CLI] Dial failed.")
import { startServer } from '../../wrappers/node/index.js';

function onKnock(sessionId, peerId) {
    console.log(`\n [SRV] New peer connected!`);
    console.log(` [SRV] Session: ${sessionId}, Peer: ${peerId}`);
}

function onMessage(sessionId, peerId, message) {
    console.log(` [SRV] Message from ${peerId}: ${message}`);
}

const serverIdentity = {
    id: "7hKot6Oqrv0BkXA_K20uOTPzAfTNAJ6Cn5Qa6IKR5NA",
    privKey: "Srnp6yXcpeoYY7glMVgfXVuFtRZmUu_zh4fYbB5qrdvuEqi3o6qu_QGRcD8rbS45M_MB9M0AnoKflBrogpHk0A",
    pubKey: "7hKot6Oqrv0BkXA_K20uOTPzAfTNAJ6Cn5Qa6IKR5NA"
};

console.log(` [NODE] Starting VSM Server on port 9999...`);
startServer(9999, serverIdentity, onKnock, onMessage);
console.log(` [NODE] Server running. Waiting for ghost traffic... (Ctrl+C to stop)`);

setInterval(() => { }, 1000);

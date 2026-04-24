import { dial, sendMessage } from '../../wrappers/node/index.js';

const clientIdentity = {
    id: "7hKot6Oqrv0BkXA_K20uOTPzAfTNAJ6Cn5Qa6IKR5NA",
    privKey: "Srnp6yXcpeoYY7glMVgfXVuFtRZmUu_zh4fYbB5qrdvuEqi3o6qu_QGRcD8rbS45M_MB9M0AnoKflBrogpHk0A",
    pubKey: "7hKot6Oqrv0BkXA_K20uOTPzAfTNAJ6Cn5Qa6IKR5NA"
};

console.log(` [NODE] Dialing 127.0.0.1:9999...`);

const sessionId = dial('lo0', '127.0.0.1', 9999, clientIdentity);

if (sessionId !== -1) {
    console.log(` [NODE] Connected! Session ID: ${sessionId}`);
    const sent = sendMessage(sessionId, 'Hello from Node.js!');
    console.log(` [NODE] Message sent: ${sent}`);
} else {
    console.log(` [NODE] Dial failed.`);
}

import { createRequire } from 'module';
import { fileURLToPath } from 'url';
import path from 'path';
import os from 'os';

const require = createRequire(import.meta.url);
const koffi = require('koffi');

const __dirname = path.dirname(fileURLToPath(import.meta.url));

// 1. Locate and load the shared library
function getLibPath() {
    const platform = os.platform();   // darwin, linux, win32
    const arch = os.arch();           // arm64, x64

    const archMap = { x64: 'amd64', arm64: 'arm64' };
    const osMap = { darwin: 'darwin', linux: 'linux', win32: 'windows' };

    const goOS = osMap[platform] || platform;
    const goArch = archMap[arch] || arch;

    const extMap = { darwin: '.dylib', linux: '.so', win32: '.dll' };
    const ext = extMap[platform] || '.so';

    // Look inside the package's own dist/ folder
    return path.resolve(__dirname, './dist', `${goOS}-${goArch}`, `vsmprotocol${ext}`);
}

const lib = koffi.load(getLibPath());

// 2. Define callback types
const KnockCallback = koffi.proto('void KnockCallback(int sessionId, const char *peerId)');
const MessageCallback = koffi.proto('void MessageCallback(int sessionId, const char *peerId, const char *message)');

// 3. Bind C functions
const _GenerateVSMIdentity = lib.func('GenerateVSMIdentity', 'void *', []);
const _FreeString = lib.func('FreeString', 'void', ['void *']);
const _StartVSMServer = lib.func('StartVSMServer', 'void', [
    'int',
    'const char *',
    koffi.pointer(KnockCallback),
    koffi.pointer(MessageCallback),
]);
const _SendMessage = lib.func('SendMessage', 'bool', ['int', 'const char *']);
const _VSMDial = lib.func('VSMDial', 'int', [
    'const char *', 'const char *', 'int', 'const char *',
]);

// 4. High-level API
const _registeredCallbacks = [];

export function generateIdentity() {
    const ptr = _GenerateVSMIdentity();
    if (!ptr) return null;
    const jsonStr = koffi.decode(ptr, 'char', -1);
    _FreeString(ptr);
    return JSON.parse(jsonStr);
}

export function startServer(port, identity, onKnock, onMessage) {
    const knockCb = koffi.register((sessionId, peerId) => {
        onKnock(sessionId, peerId);
    }, koffi.pointer(KnockCallback));

    const msgCb = koffi.register((sessionId, peerId, message) => {
        onMessage(sessionId, peerId, message);
    }, koffi.pointer(MessageCallback));

    _registeredCallbacks.push(knockCb, msgCb);

    _StartVSMServer(port, JSON.stringify(identity), knockCb, msgCb);
}

export function dial(device, targetIP, targetPort, identity) {
    return _VSMDial(device, targetIP, targetPort, JSON.stringify(identity));
}

export function sendMessage(sessionId, text) {
    return _SendMessage(sessionId, text);
}

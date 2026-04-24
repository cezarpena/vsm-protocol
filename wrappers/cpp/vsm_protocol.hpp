#pragma once

#include <string>
#include <functional>
#include <cstdlib>
#include <dlfcn.h>
#include <stdexcept>

namespace vsm {

// C function pointer types (matching the .dylib exports)
using KnockCallbackFn   = void (*)(int, char*);
using MessageCallbackFn = void (*)(int, char*, char*);

// C function signatures
using GenerateIdentityFn = char* (*)();
using FreeStringFn       = void (*)(char*);
using StartServerFn      = void (*)(long long, char*, KnockCallbackFn, MessageCallbackFn);
using SendMessageFn      = unsigned char (*)(long long, char*);
using DialFn             = long long (*)(char*, char*, long long, char*);

// Global callback storage (prevents dangling pointers)
inline std::function<void(int, std::string)> g_onKnock;
inline std::function<void(int, std::string, std::string)> g_onMessage;

// C-compatible trampolines
inline void knockTrampoline(int sessionId, char* peerId) {
    if (g_onKnock) g_onKnock(sessionId, std::string(peerId));
}
inline void messageTrampoline(int sessionId, char* peerId, char* msg) {
    if (g_onMessage) g_onMessage(sessionId, std::string(peerId), std::string(msg));
}

class Protocol {
public:
    explicit Protocol(const std::string& libPath) {
        handle_ = dlopen(libPath.c_str(), RTLD_LAZY);
        if (!handle_) throw std::runtime_error(std::string("Failed to load library: ") + dlerror());

        generateIdentity_ = (GenerateIdentityFn)dlsym(handle_, "GenerateVSMIdentity");
        freeString_       = (FreeStringFn)dlsym(handle_, "FreeString");
        startServer_      = (StartServerFn)dlsym(handle_, "StartVSMServer");
        sendMessage_      = (SendMessageFn)dlsym(handle_, "SendMessage");
        dial_             = (DialFn)dlsym(handle_, "VSMDial");
    }

    ~Protocol() {
        if (handle_) dlclose(handle_);
    }

    // Non-copyable
    Protocol(const Protocol&) = delete;
    Protocol& operator=(const Protocol&) = delete;

    /// Generate a new JSON identity string
    std::string generateIdentity() {
        char* ptr = generateIdentity_();
        if (!ptr) return "";
        std::string result(ptr);
        freeString_(ptr);
        return result;
    }

    /// Start a stealth server on the given port
    void startServer(int port, const std::string& identityJson,
                     std::function<void(int, std::string)> onKnock,
                     std::function<void(int, std::string, std::string)> onMessage) {
        g_onKnock = std::move(onKnock);
        g_onMessage = std::move(onMessage);
        startServer_(port, const_cast<char*>(identityJson.c_str()),
                     knockTrampoline, messageTrampoline);
    }

    /// Dial a remote VSM server. Returns session ID or -1 on failure.
    int dial(const std::string& device, const std::string& targetIP,
             int port, const std::string& identityJson) {
        return static_cast<int>(dial_(
            const_cast<char*>(device.c_str()),
            const_cast<char*>(targetIP.c_str()),
            port,
            const_cast<char*>(identityJson.c_str())
        ));
    }

    /// Send an encrypted message on an active session
    bool sendMessage(int sessionId, const std::string& text) {
        return sendMessage_(sessionId, const_cast<char*>(text.c_str())) != 0;
    }

private:
    void* handle_ = nullptr;
    GenerateIdentityFn generateIdentity_ = nullptr;
    FreeStringFn       freeString_ = nullptr;
    StartServerFn      startServer_ = nullptr;
    SendMessageFn      sendMessage_ = nullptr;
    DialFn             dial_ = nullptr;
};

} // namespace vsm

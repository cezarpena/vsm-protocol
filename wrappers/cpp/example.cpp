#include "vsm_protocol.hpp"
#include <iostream>
#include <thread>
#include <chrono>

int main() {
    try {
        vsm::Protocol vsm("./vsmprotocol.dylib");

        std::string identity = vsm.generateIdentity();
        std::cout << " [C++] Identity: " << identity << std::endl;

        // Example: Start a server
        // vsm.startServer(9999, identity,
        //     [](int sid, std::string peer) {
        //         std::cout << " [C++] Knock from " << peer << std::endl;
        //     },
        //     [](int sid, std::string peer, std::string msg) {
        //         std::cout << " [C++] Message: " << msg << std::endl;
        //     }
        // );

        // Example: Dial
        // int sid = vsm.dial("lo0", "127.0.0.1", 9999, identity);
        // if (sid != -1) vsm.sendMessage(sid, "Hello from C++!");

        std::cout << " [C++] VSM Protocol loaded successfully." << std::endl;

    } catch (const std::exception& e) {
        std::cerr << "Error: " << e.what() << std::endl;
        return 1;
    }
    return 0;
}

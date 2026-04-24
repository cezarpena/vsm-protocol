package vsm;

import com.sun.jna.*;
import com.sun.jna.Callback;

/**
 * VSM Stealth Protocol - Java/JVM Wrapper
 * 
 * Uses JNA to call the Go-compiled shared library.
 * Works with Java, Kotlin, Scala, Clojure, or any JVM language.
 *
 * Dependency (Maven):
 *   <dependency>
 *     <groupId>net.java.dev.jna</groupId>
 *     <artifactId>jna</artifactId>
 *     <version>5.14.0</version>
 *   </dependency>
 *
 * Dependency (Gradle):
 *   implementation 'net.java.dev.jna:jna:5.14.0'
 */
public class VSMProtocol {

    // --- Callback interfaces ---
    public interface KnockCallback extends Callback {
        void invoke(int sessionId, String peerId);
    }

    public interface MessageCallback extends Callback {
        void invoke(int sessionId, String peerId, String message);
    }

    // --- Native library binding ---
    public interface VSMLib extends Library {
        Pointer GenerateVSMIdentity();
        void FreeString(Pointer str);
        void StartVSMServer(long port, String idStr, KnockCallback knockCb, MessageCallback msgCb);
        byte SendMessage(long sessionId, String message);
        long VSMDial(String device, String targetIP, long port, String idStr);
    }

    private final VSMLib lib;

    // Must hold strong references to callbacks to prevent GC
    private KnockCallback knockRef;
    private MessageCallback messageRef;

    /**
     * Load the VSM Protocol library.
     * @param libPath Path to vsmprotocol.dylib / .so / .dll
     */
    public VSMProtocol(String libPath) {
        this.lib = Native.load(libPath, VSMLib.class);
    }

    /**
     * Generate a new identity as a JSON string.
     */
    public String generateIdentity() {
        Pointer ptr = lib.GenerateVSMIdentity();
        if (ptr == null) return null;
        String json = ptr.getString(0);
        lib.FreeString(ptr);
        return json;
    }

    /**
     * Start a stealth server on the given port.
     */
    public void startServer(int port, String identityJson,
                            KnockCallback onKnock, MessageCallback onMessage) {
        // Hold strong references so GC doesn't collect them
        this.knockRef = onKnock;
        this.messageRef = onMessage;
        lib.StartVSMServer(port, identityJson, onKnock, onMessage);
    }

    /**
     * Dial a remote VSM server.
     * @return Session ID on success, -1 on failure
     */
    public int dial(String device, String targetIP, int port, String identityJson) {
        return (int) lib.VSMDial(device, targetIP, port, identityJson);
    }

    /**
     * Send an encrypted message over an active session.
     */
    public boolean sendMessage(int sessionId, String text) {
        return lib.SendMessage(sessionId, text) != 0;
    }

    // --- Demo ---
    public static void main(String[] args) throws Exception {
        VSMProtocol vsm = new VSMProtocol("./vsmprotocol.dylib");

        String identity = vsm.generateIdentity();
        System.out.println(" [JAVA] Identity: " + identity);

        // Server example:
        // vsm.startServer(9999, identity,
        //     (sid, peer) -> System.out.println(" [JAVA] Knock from " + peer),
        //     (sid, peer, msg) -> System.out.println(" [JAVA] Message: " + msg)
        // );

        // Client example:
        // int sid = vsm.dial("lo0", "127.0.0.1", 9999, identity);
        // if (sid != -1) vsm.sendMessage(sid, "Hello from Java!");

        System.out.println(" [JAVA] VSM Protocol loaded successfully.");
    }
}

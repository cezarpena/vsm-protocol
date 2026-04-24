using System;
using System.Runtime.InteropServices;

namespace VSM
{
    /// <summary>
    /// VSM Stealth Protocol - C# / .NET Wrapper
    /// 
    /// Uses P/Invoke to call the Go-compiled shared library.
    /// Works with .NET 6+, .NET Framework 4.7+, Mono, and Unity.
    ///
    /// Usage:
    ///   dotnet new console -n VsmTest
    ///   copy this file into the project
    ///   dotnet run
    /// </summary>
    public static class VSMProtocol
    {
        // Update this to match your platform and path
        private const string LibName = "./vsmprotocol.dylib";

        // --- Callback delegates ---
        [UnmanagedFunctionPointer(CallingConvention.Cdecl)]
        public delegate void KnockCallback(int sessionId, IntPtr peerId);

        [UnmanagedFunctionPointer(CallingConvention.Cdecl)]
        public delegate void MessageCallback(int sessionId, IntPtr peerId, IntPtr message);

        // --- P/Invoke bindings ---
        [DllImport(LibName, CallingConvention = CallingConvention.Cdecl)]
        private static extern IntPtr GenerateVSMIdentity();

        [DllImport(LibName, CallingConvention = CallingConvention.Cdecl)]
        private static extern void FreeString(IntPtr str);

        [DllImport(LibName, CallingConvention = CallingConvention.Cdecl)]
        private static extern void StartVSMServer(
            long port,
            [MarshalAs(UnmanagedType.LPUTF8Str)] string idStr,
            KnockCallback knockCb,
            MessageCallback msgCb);

        [DllImport(LibName, CallingConvention = CallingConvention.Cdecl)]
        private static extern byte SendMessage(
            long sessionId,
            [MarshalAs(UnmanagedType.LPUTF8Str)] string message);

        [DllImport(LibName, CallingConvention = CallingConvention.Cdecl)]
        private static extern long VSMDial(
            [MarshalAs(UnmanagedType.LPUTF8Str)] string device,
            [MarshalAs(UnmanagedType.LPUTF8Str)] string targetIP,
            long port,
            [MarshalAs(UnmanagedType.LPUTF8Str)] string idStr);

        // --- prevent GC of callback delegates ---
        private static KnockCallback? _knockRef;
        private static MessageCallback? _messageRef;

        // --- High-level API ---

        /// <summary>Generate a new VSM identity as a JSON string.</summary>
        public static string? GenerateIdentity()
        {
            IntPtr ptr = GenerateVSMIdentity();
            if (ptr == IntPtr.Zero) return null;
            string json = Marshal.PtrToStringUTF8(ptr)!;
            FreeString(ptr);
            return json;
        }

        /// <summary>Start a stealth server on the given port.</summary>
        public static void StartServer(int port, string identityJson,
            Action<int, string> onKnock, Action<int, string, string> onMessage)
        {
            _knockRef = (sid, peerPtr) =>
            {
                string peer = Marshal.PtrToStringUTF8(peerPtr) ?? "";
                onKnock(sid, peer);
            };

            _messageRef = (sid, peerPtr, msgPtr) =>
            {
                string peer = Marshal.PtrToStringUTF8(peerPtr) ?? "";
                string msg = Marshal.PtrToStringUTF8(msgPtr) ?? "";
                onMessage(sid, peer, msg);
            };

            StartVSMServer(port, identityJson, _knockRef, _messageRef);
        }

        /// <summary>Dial a remote VSM server. Returns session ID or -1.</summary>
        public static int Dial(string device, string targetIP, int port, string identityJson)
        {
            return (int)VSMDial(device, targetIP, port, identityJson);
        }

        /// <summary>Send an encrypted message over an active session.</summary>
        public static bool Send(int sessionId, string text)
        {
            return SendMessage(sessionId, text) != 0;
        }
    }

    // --- Demo ---
    class Program
    {
        static void Main(string[] args)
        {
            string? identity = VSMProtocol.GenerateIdentity();
            Console.WriteLine($" [C#] Identity: {identity}");

            // Server example:
            // VSMProtocol.StartServer(9999, identity!,
            //     (sid, peer) => Console.WriteLine($" [C#] Knock from {peer}"),
            //     (sid, peer, msg) => Console.WriteLine($" [C#] Message: {msg}")
            // );

            // Client example:
            // int sid = VSMProtocol.Dial("lo0", "127.0.0.1", 9999, identity!);
            // if (sid != -1) VSMProtocol.Send(sid, "Hello from C#!");

            Console.WriteLine(" [C#] VSM Protocol loaded successfully.");
        }
    }
}

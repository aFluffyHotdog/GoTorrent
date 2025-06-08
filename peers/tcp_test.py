# Save as tcp_connect_test.py

import socket
import sys

def test_tcp_connect(ip, port, timeout=5):
    try:
        with socket.create_connection((ip, port), timeout=timeout) as sock:
            print(f"Successfully connected to {ip}:{port}")
    except Exception as e:
        print(f"Failed to connect to {ip}:{port} - {e}")

if __name__ == "__main__":
    if len(sys.argv) != 3:
        print("Usage: python tcp_connect_test.py <ip> <port>")
        sys.exit(1)
    ip = sys.argv[1]
    port = int(sys.argv[2])
    test_tcp_connect(ip, port)
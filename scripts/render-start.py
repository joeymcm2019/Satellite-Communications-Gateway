import os
import signal
import subprocess
import sys
import time

processes = []


def shutdown(_signum, _frame):
    for process in processes:
        if process.poll() is None:
            process.terminate()


signal.signal(signal.SIGTERM, shutdown)
signal.signal(signal.SIGINT, shutdown)

try:
    gateway = subprocess.Popen(["/usr/local/bin/gateway"])
    processes.append(gateway)

    simulator = subprocess.Popen(
        [sys.executable, "-m", "satellite_simulator"],
        env={
            **os.environ,
            "GATEWAY_GRPC_ADDRESS": "127.0.0.1:50051",
        },
    )
    processes.append(simulator)

    while all(process.poll() is None for process in processes):
        time.sleep(0.5)

finally:
    shutdown(None, None)

    for process in processes:
        try:
            process.wait(timeout=5)
        except subprocess.TimeoutExpired:
            process.kill()
            process.wait()

sys.exit(1)

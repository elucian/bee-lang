import os
import time
import subprocess
import json

def run_benchmarks():
    print("Running Bee Self-Testing & Performance Suite...")
    start_time = time.time()
    
    bench_file = "test/bench/collection_bench.bee"
    t_start = time.perf_counter()
    res = subprocess.run(["go", "run", "cmd/bee/main.go", "-e", bench_file], capture_output=True, text=True)
    t_end = time.perf_counter()
    
    elapsed_ms = (t_end - t_start) * 1000.0
    success = res.returncode == 0
    
    report = {
        "timestamp": time.strftime("%Y-%m-%d %H:%M:%S"),
        "benchmark": bench_file,
        "success": success,
        "execution_time_ms": elapsed_ms,
        "stdout": res.stdout.strip(),
        "stderr": res.stderr.strip()
    }
    
    os.makedirs("test/output", exist_ok=True)
    report_path = "test/output/perf_report.json"
    with open(report_path, "w") as f:
        json.dump(report, f, indent=2)
        
    print(f"Performance report generated at {report_path} (Time: {elapsed_ms:.2f} ms)")

if __name__ == "__main__":
    run_benchmarks()

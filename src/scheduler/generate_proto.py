import os
import subprocess

def generate_proto_files():
    """Generate Python protobuf files from trade.proto"""
    # Get the current directory
    current_dir = os.path.dirname(os.path.abspath(__file__))

    # Path to the proto file
    proto_file = os.path.join(current_dir, 'tradepb', 'trade.proto')

    # Path to the output directory (utils directory)
    output_dir = os.path.join(current_dir, 'strategies', 'utils')

    # Ensure the output directory exists
    os.makedirs(output_dir, exist_ok=True)

    # Command to generate Python protobuf files
    venv_python = os.path.join(os.path.dirname(os.path.dirname(current_dir)), '.venv', 'Scripts', 'python.exe')
    cmd = [
        venv_python, '-m', 'grpc_tools.protoc',
        f'--proto_path={os.path.dirname(proto_file)}',
        f'--python_out={output_dir}',
        f'--grpc_python_out={output_dir}',
        proto_file
    ]
    print(f"Using Python: {venv_python}")
    print(f"Proto file: {proto_file}")
    print(f"Output directory: {output_dir}")

    # Run the command
    print(f"Running command: {' '.join(cmd)}")
    result = subprocess.run(cmd, capture_output=True, text=True)

    if result.returncode != 0:
        print(f"Error generating protobuf files: {result.stderr}")
        return False

    print(f"Command output: {result.stdout}")

    print("Successfully generated protobuf files:")
    print(f"  - {os.path.join(output_dir, 'trade_pb2.py')}")
    print(f"  - {os.path.join(output_dir, 'trade_pb2_grpc.py')}")

    # Fix imports in the generated files
    fix_imports(os.path.join(output_dir, 'trade_pb2_grpc.py'))

    return True

def fix_imports(grpc_file):
    """Fix imports in the generated grpc file"""
    with open(grpc_file, 'r') as f:
        content = f.read()

    # Replace the import statement
    content = content.replace(
        'import trade_pb2 as trade__pb2',
        'import utils.trade_pb2 as trade__pb2'
    )

    with open(grpc_file, 'w') as f:
        f.write(content)

    print(f"Fixed imports in {grpc_file}")

if __name__ == "__main__":
    generate_proto_files()

import sys
from pptx import Presentation
import json
import traceback
from contextlib import redirect_stdout
import io

def run_repl(code_str: str, file_path: str = None):
    # This acts as a sandbox to execute Agent-generated Python code 
    # manipulating the given PPTX using python-pptx.
    
    # Pre-load presentation if provided
    prs = None
    if file_path:
        prs = Presentation(file_path)
    
    # Environment variables for sandbox
    env = {
        "Presentation": Presentation,
        "prs": prs,
        "json": json
    }
    
    output_buf = io.StringIO()
    try:
        with redirect_stdout(output_buf):
            # Execute generated agent code
            exec(code_str, env)
        
        # Determine if we should save the edited presentation
        if prs and file_path:
            out_file = file_path.replace('.pptx', '_edited.pptx')
            prs.save(out_file)
            
        return {"status": "success", "output": output_buf.getvalue(), "error": None}
    except Exception as e:
        return {"status": "error", "output": output_buf.getvalue(), "error": traceback.format_exc()}

if __name__ == "__main__":
    import argparse
    parser = argparse.ArgumentParser()
    parser.add_argument("--code", type=str, required=True, help="Python code to run")
    parser.add_argument("--ppt", type=str, required=False, help="PPT file to edit")
    args = parser.parse_args()
    
    result = run_repl(args.code, args.ppt)
    print(json.dumps(result))
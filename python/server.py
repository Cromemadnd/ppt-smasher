from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import sys
import io
import traceback
from contextlib import redirect_stdout
from typing import Optional
from pptx import Presentation

app = FastAPI(title="PPT Smasher Agent REPL Sandbox")

class ExecRequest(BaseModel):
    code: str
    ppt_path: Optional[str] = None

class ExecResponse(BaseModel):
    status: str
    output: str
    error: Optional[str] = None

@app.post("/api/repl", response_model=ExecResponse)
def execute_agent_code(req: ExecRequest):
    # This REPL sandbox accepts Python code from the Script Coder Agent
    # and executes it safely, injecting `python-pptx` dependencies.
    
    prs = None
    if req.ppt_path:
        try:
            prs = Presentation(req.ppt_path)
        except Exception as e:
            raise HTTPException(status_code=400, detail=f"Failed to load PPT: {e}")
            
    # Prepare global execution environment
    env = {
        "Presentation": Presentation,
        "prs": prs,
    }

    output_buf = io.StringIO()
    try:
        with redirect_stdout(output_buf):
            # Run the agent's code in the sandbox env
            exec(req.code, env)
        
        # Save if edited
        if prs and req.ppt_path:
            save_path = req.ppt_path.replace(".pptx", "_agent_edited.pptx")
            prs.save(save_path)
            
        return ExecResponse(status="success", output=output_buf.getvalue())
    except Exception as e:
        # Pass traceback directly to the agent to fix logic if crashed
        error_trace = traceback.format_exc()
        return ExecResponse(status="error", output=output_buf.getvalue(), error=error_trace)

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)

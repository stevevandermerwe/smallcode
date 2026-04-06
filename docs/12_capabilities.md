Building successful production agents is roughly **80% infrastructure engineering** and only **20% AI modeling**. The primary value of an agentic system lies in its architectural primitives rather than the underlying LLM.

The following 12 points define the essential engineering requirements for moving from a prototype to an enterprise-grade agent:

### Core Architectural Principles
1.  **Prioritize Infrastructure over Models:** Real-world utility is driven by the "boring" plumbing—the primitives that support the model—not the model itself.
2.  **The 80/20 Execution Rule:** Success in production depends on rigorous engineering of non-glamorous components (security, state, and logging) rather than optimizing prompts or model weights.
3.  **Avoid Premature Complexity:** Focus on foundational engineering principles first. Over-engineering early-stage agent logic often leads to failure modes that are difficult to debug.

### Infrastructure Primitives
4.  **Metadata-First Tool Registry:** Maintain a registry that defines agent capabilities and schemas without executing code. This allows the system to manage and inspect agent possibilities safely.
5.  **Tiered Permission Architecture:** Implement multi-layered security protocols (such as 18-module architectures for high-risk tools like Bash) to establish granular trust levels.
6.  **Full Session Persistence:** Ensure the system can reconstruct the entire state—including conversation history, token usage, and active permissions—immediately following a system crash.
7.  **Decoupled Workflow State:** Separate conversational history from the functional task state. This ensures that workflows can be retried or resumed after failures without polluting the dialogue context.
8.  **Proactive Token Budgeting:** Integrate pre-turn checks and active budget tracking to prevent runaway costs and infinite loops before they occur.
9.  **Structured System Observability:** Develop deep, structured streaming events and logs. This is critical for transparency and forensic reconstruction when an agent deviates from expected behavior.
10. **Harness-Level Verification:** Use a dual-testing approach combining task-specific validation with broad harness testing to ensure updates don't compromise security guardrails.

### Operational Maturity
11. **Dynamic Tool Assembly:** Instead of loading an entire library, dynamically assemble "tool pools" based on the specific requirements of the current session to reduce overhead and model confusion.
12. **Context Compaction & Role Specialization:** Automate the removal of non-essential transcript context to manage token windows, and organize agents into specific functional roles (e.g., Planner, Executor, Verifier) to maintain strict control over system behavior.

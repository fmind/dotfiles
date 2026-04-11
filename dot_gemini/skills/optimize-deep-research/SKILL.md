---
name: optimize-deep-research
description: Transform vague research ideas into high-quality, structured prompts for Gemini's deep research mode.
---

# Optimize Deep Research

## Overview

This skill helps you frame complex research questions by expanding your initial idea into a structured prompt that covers persona, scope, specific questions, and output requirements.

## Workflow

1. **Analyze & Expand**: When a user provides a research idea, identify its core objective and provide an initial "Deep Research Prompt" that includes:
   - **Role**: A specific expert persona (e.g., "Senior Market Analyst", "Technical Researcher").
   - **Objective**: A clear statement of what needs to be discovered.
   - **Scope**: What is in-scope vs. out-of-scope.
   - **Key Questions**: 3-5 specific, open-ended questions to guide the research.
   - **Output Strategy**:
     - **Executive Summary**: Always placed at the very top.
     - **Limitations & Challenges**: A dedicated section explaining the constraints of the research and what remains unknown.
     - **Definitive Recommendation**: The research must deliver an opinionated conclusion and a clear, actionable recommendation.

2. **Refine**: After providing the initial draft, ask **2-3 targeted questions** to help the user narrow down the focus (e.g., "Which region?", "What timeframe?", "Technical vs. Economic priority?").

3. **Finalize**: Once the user provides feedback, generate the final, copy-pasteable prompt.

## Example Prompts & Questions

### General Research Ideas

- **Idea**: "Research the future of AI in healthcare."
- **Refinement Questions**: "Are you focusing on diagnostic tools or administrative efficiency?", "What is the specific geographic focus?", "Are you looking at the next 2 years or the next decade?"

### Technical/Strategic Ideas

- **Idea**: "Compare different cloud migration strategies."
- **Refinement Questions**: "Is this for a small startup or a large enterprise?", "What are the primary cost constraints?", "Are there specific cloud providers (AWS, Azure, GCP) to prioritize?"

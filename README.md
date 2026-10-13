# VAC-war

Research repository documenting hypotheses, observations, and testing related to reported changes in Counter-Strike 2's Valve Anti-Cheat (VAC) behavior.

## Project Goals

- Document hypotheses concerning reported VAC behavior changes.
- Separate verified information from speculation.
- Maintain reproducible test plans and structured observations.
- Evaluate hypotheses against evidence rather than assumptions.

## Repository Structure

```text
VAC-war/
├── README.md
├── research/
│   ├── pre-update/
│   │   └── first-theory/
│   │       └── theory-1.md
│   └── post-update/
│       └── first-theory/
│           ├── theory-1.md
│           └── keywords.md
└── tests/
    └── post-update/
        └── test-1/
            ├── README.md
            ├── test-plan.md
            ├── test-environment.md
            └── results.md
```

## Research Registry

| Research | Description | Status |
|---|---|---|
| [Pre-Update Theory 1](research/pre-update/first-theory/theory-1.md) | Preliminary overview of possible VAC detection mechanisms | Preliminary |
| [Post-Update Theory 1](research/post-update/first-theory/theory-1.md) | Hypothesis concerning process-name inspection | Unverified |
| [Hypothesized Keywords](research/post-update/first-theory/keywords.md) | Exploratory list of potentially relevant terms | Unverified |

## Test Registry

| Test | Related Research | Status |
|---|---|---|
| [Post-Update Test 1](tests/post-update/test-1/README.md) | [Post-Update Theory 1](research/post-update/first-theory/theory-1.md) | Planned |

## Current Project Status

**Current Stage:** Stage 1 — Research and Analysis

**Overall Status:** Ongoing

**Phase:** Test Preparation

The first post-update experiment is planned for October 17–25, 2026. The proposed sample is 15 matches. Testing has not started, and the hypothesis remains unverified.

### Preparation Checklist

- [ ] Finalize the test plan
- [ ] Document the test environment
- [ ] Verify the test environment
- [ ] Conduct the experiment
- [ ] Record observations
- [ ] Evaluate the results

Update these checkboxes only when each task has actually been completed.

## Research Methodology

Each research document should distinguish between:

- **Verified information:** Supported by reliable sources or reproducible evidence.
- **Hypothesis:** A proposed explanation that has not been established.
- **Observation:** Something directly recorded during research or testing.
- **Interpretation:** An explanation of an observation.
- **Limitation:** A factor preventing a definitive conclusion.

A lack of visible detection or enforcement does not establish that VAC did not detect something internally.

A ban or other enforcement event does not, by itself, identify the mechanism responsible.

## Evidence Standards

- Cite sources for technical claims wherever possible.
- Record relevant versions, dates, and conditions.
- Distinguish correlation from causation.
- Document contradictory observations.
- Avoid treating community reports as independently verified facts.
- Do not claim that a hypothesis is confirmed without evidence that directly supports it.

## Sources

- [Counter-Strike 2 Official Updates](https://store.steampowered.com/news/app/730)
- [CS2 VAC Internals — community research](https://github.com/Aspasia1337/cs2-vac-internals)
- [GitHub search: VAC anti-cheat research](https://github.com/search?q=VAC+anti-cheat+reverse+engineering&type=repositories)

Sources provide research context and do not necessarily validate every claim in this repository.

## Disclaimer

This repository is an independent research project. Hypotheses and observations are not official statements from Valve. No claim about VAC's internal implementation should be considered established solely because it appears here.

Testing should be conducted only in authorized environments and in accordance with applicable rules and terms.

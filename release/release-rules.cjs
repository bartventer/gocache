/**
 * Custom `releaseRules` rules before v1.0.0.
 * For the original rules, see: https://github.com/semantic-release/commit-analyzer/blob/master/lib/default-release-rules.js
 *
 * @type {Array<{
*   breaking?: boolean;
*   revert?: boolean;
*   type?: "feat" | "fix" | "perf" | "chore" | "docs" | "style" | "refactor" | "test" | "build" | "ci" | "revert";
*   scope?: string;
*   emoji?: string;
*   tag?: string;
*   component?: string;
*   release: "patch" | "minor" | "major";
* }>}
*/
module.exports = [
    { breaking: true, release: "minor" }, // Custom rule for breaking changes
    { revert: true, release: "patch" },
    // Angular
    { type: "feat", release: "minor" },
    { type: "fix", release: "patch" },
    { type: "perf", release: "patch" },
    // Atom
    { emoji: ":racehorse:", release: "patch" },
    { emoji: ":bug:", release: "patch" },
    { emoji: ":penguin:", release: "patch" },
    { emoji: ":apple:", release: "patch" },
    { emoji: ":checkered_flag:", release: "patch" },
    // Ember
    { tag: "BUGFIX", release: "patch" },
    { tag: "FEATURE", release: "minor" },
    { tag: "SECURITY", release: "patch" },
    // ESLint
    { tag: "Breaking", release: "minor" }, // Custom rule for breaking changes
    { tag: "Fix", release: "patch" },
    { tag: "Update", release: "minor" },
    { tag: "New", release: "minor" },
    // Express
    { component: "perf", release: "patch" },
    { component: "deps", release: "patch" },
    // JSHint
    { type: "FEAT", release: "minor" },
    { type: "FIX", release: "patch" },
];
import { useState } from "react";
import type { FormEvent } from "react";
import type { Project } from "../types";

type ProjectData = Omit<Project, "id" | "created_at">;

interface ProjectFormProps {
  initial?: Project;
  onSubmit: (data: ProjectData) => Promise<void>;
  onCancel: () => void;
}

const inputClass =
  "w-full bg-[#0f3460]/50 border border-[#e2b714]/10 rounded px-3 py-2.5 text-white text-sm placeholder-[#a0a0b8]/40 focus:outline-none focus:border-[#e2b714]/40 focus:ring-1 focus:ring-[#e2b714]/20 transition-colors font-mono";

const labelClass =
  "block text-[#a0a0b8] text-xs font-medium uppercase tracking-wider mb-1.5";

export default function ProjectForm({
  initial,
  onSubmit,
  onCancel,
}: ProjectFormProps) {
  const [name, setName] = useState(initial?.name ?? "");
  const [path, setPath] = useState(initial?.path ?? "");
  const [description, setDescription] = useState(initial?.description ?? "");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");

  const nameValid = name.length >= 1 && name.length <= 100;
  const pathValid = path.length >= 1;
  const canSubmit = nameValid && pathValid && !submitting;

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    if (!canSubmit) return;
    setError("");
    setSubmitting(true);
    try {
      await onSubmit({ name, path, description });
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : "Something went wrong";
      setError(msg);
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-5">
      {/* Header */}
      <div className="flex items-center gap-3 pb-4 border-b border-[#e2b714]/10">
        <div className="w-8 h-8 rounded bg-[#e2b714]/10 border border-[#e2b714]/25 flex items-center justify-center">
          <span className="text-[#e2b714] text-sm font-bold">
            {name.charAt(0).toUpperCase() || "?"}
          </span>
        </div>
        <div>
          <h2 className="text-white text-base font-bold font-mono">
            {initial ? "Edit Project" : "New Project"}
          </h2>
          <p className="text-[#a0a0b8]/50 text-xs font-mono">
            {initial
              ? `Editing ${initial.name}`
              : "Register a project for agent context"}
          </p>
        </div>
      </div>

      {error && (
        <div
          role="alert"
          className="bg-red-500/10 border border-red-500/20 text-red-400 text-sm px-3 py-2 rounded font-mono"
        >
          {error}
        </div>
      )}

      {/* Name */}
      <div>
        <label htmlFor="project-name" className={labelClass}>
          Name
        </label>
        <input
          id="project-name"
          type="text"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="e.g. waynebot"
          maxLength={100}
          className={inputClass}
        />
        {name.length > 0 && !nameValid && (
          <p className="text-red-400/80 text-xs mt-1 font-mono">
            Name must be 1-100 characters
          </p>
        )}
      </div>

      {/* Path */}
      <div>
        <label htmlFor="project-path" className={labelClass}>
          Path
        </label>
        <input
          id="project-path"
          type="text"
          value={path}
          onChange={(e) => setPath(e.target.value)}
          placeholder="e.g. /home/user/projects/waynebot"
          className={inputClass}
        />
        {path.length === 0 && name.length > 0 && (
          <p className="text-red-400/80 text-xs mt-1 font-mono">
            Path is required
          </p>
        )}
      </div>

      {/* Description */}
      <div>
        <label htmlFor="project-description" className={labelClass}>
          Description{" "}
          <span className="text-[#a0a0b8]/30 normal-case tracking-normal">
            optional
          </span>
        </label>
        <textarea
          id="project-description"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          placeholder="Brief description of the project..."
          rows={3}
          className={`${inputClass} resize-y min-h-[80px]`}
        />
      </div>

      {/* Actions */}
      <div className="flex items-center gap-3 pt-4 border-t border-[#e2b714]/10">
        <button
          type="submit"
          disabled={!canSubmit}
          className="bg-[#e2b714] hover:bg-[#c9a212] disabled:bg-[#e2b714]/20 disabled:text-[#a0a0b8]/40 text-[#1a1a2e] font-semibold text-sm py-2.5 px-5 rounded transition-colors cursor-pointer disabled:cursor-not-allowed"
        >
          {submitting
            ? "Saving..."
            : initial
              ? "Update Project"
              : "Create Project"}
        </button>
        <button
          type="button"
          onClick={onCancel}
          className="text-[#a0a0b8]/60 hover:text-[#a0a0b8] text-sm py-2.5 px-4 rounded border border-[#e2b714]/10 hover:border-[#e2b714]/25 transition-colors cursor-pointer"
        >
          Cancel
        </button>
      </div>
    </form>
  );
}

import { useEffect, useState } from "react";
import type { FormEvent } from "react";
import type { Persona, PersonaTemplate } from "../types";
import { getPersonaTemplates } from "../api";
import { getErrorMessage } from "../utils/errors";
import { inputClass, labelClass } from "../utils/styles";

type PersonaData = Omit<Persona, "id" | "created_at">;

interface PersonaFormProps {
  initial?: Persona;
  onSubmit: (data: PersonaData) => Promise<void>;
  onCancel: () => void;
}

export default function PersonaForm({
  initial,
  onSubmit,
  onCancel,
}: PersonaFormProps) {
  const [name, setName] = useState(initial?.name ?? "");
  const [systemPrompt, setSystemPrompt] = useState(
    initial?.system_prompt ?? "",
  );
  const [model, setModel] = useState(initial?.model ?? "");
  const [temperature, setTemperature] = useState(initial?.temperature ?? 0.7);
  const [maxTokens, setMaxTokens] = useState(initial?.max_tokens ?? 4096);
  const [cooldownSecs, setCooldownSecs] = useState(initial?.cooldown_secs ?? 0);
  const [maxTokensPerHour, setMaxTokensPerHour] = useState(
    initial?.max_tokens_per_hour ?? 100000,
  );
  const [toolsRaw, setToolsRaw] = useState(
    initial?.tools_enabled.join(", ") ?? "",
  );
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");

  const [templates, setTemplates] = useState<PersonaTemplate[]>([]);

  useEffect(() => {
    if (!initial) {
      getPersonaTemplates()
        .then(setTemplates)
        .catch(() => {});
    }
  }, [initial]);

  function applyTemplate(t: PersonaTemplate) {
    setName(t.name);
    setSystemPrompt(t.system_prompt);
    setModel(t.model);
    setTemperature(t.temperature);
    setMaxTokens(t.max_tokens);
    setCooldownSecs(t.cooldown_secs);
    setToolsRaw(t.tools_enabled.join(", "));
  }

  const nameValid = name.length >= 1 && name.length <= 100;
  const promptValid = systemPrompt.length >= 1 && systemPrompt.length <= 50000;
  const modelValid = model.length > 0;
  const canSubmit = nameValid && promptValid && modelValid && !submitting;

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    if (!canSubmit) return;
    setError("");
    setSubmitting(true);
    try {
      const toolsEnabled = toolsRaw
        .split(",")
        .map((t) => t.trim())
        .filter(Boolean);
      await onSubmit({
        name,
        system_prompt: systemPrompt,
        model,
        temperature,
        max_tokens: maxTokens,
        cooldown_secs: cooldownSecs,
        max_tokens_per_hour: maxTokensPerHour,
        tools_enabled: toolsEnabled,
      });
    } catch (err: unknown) {
      setError(getErrorMessage(err));
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-5">
      {/* Header */}
      <div className="flex items-center gap-3 pb-4 border-b border-[#e2b714]/10">
        <div className="w-8 h-8 rotate-45 bg-[#e2b714]/10 border border-[#e2b714]/25 flex items-center justify-center">
          <span className="-rotate-45 text-[#e2b714] text-sm font-bold">
            {name.charAt(0).toUpperCase() || "?"}
          </span>
        </div>
        <div>
          <h2 className="text-white text-base font-bold font-mono">
            {initial ? "Edit Persona" : "New Persona"}
          </h2>
          <p className="text-[#a0a0b8]/50 text-xs font-mono">
            {initial
              ? `Editing ${initial.name}`
              : "Configure an AI agent persona"}
          </p>
        </div>
      </div>

      {/* Template selector — only for new personas */}
      {!initial && templates.length > 0 && (
        <div>
          <label className={labelClass}>Start from template</label>
          <div className="flex flex-wrap gap-2">
            {templates.map((t) => (
              <button
                key={t.name}
                type="button"
                onClick={() => applyTemplate(t)}
                className={`text-xs px-3 py-1.5 rounded border transition-colors cursor-pointer font-mono ${
                  name === t.name
                    ? "bg-[#e2b714]/20 border-[#e2b714]/40 text-[#e2b714]"
                    : "bg-[#0f3460]/30 border-[#e2b714]/10 text-[#a0a0b8]/70 hover:border-[#e2b714]/25 hover:text-[#a0a0b8]"
                }`}
              >
                {t.name}
              </button>
            ))}
          </div>
        </div>
      )}

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
        <label htmlFor="persona-name" className={labelClass}>
          Name
        </label>
        <input
          id="persona-name"
          type="text"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="e.g. helpful-assistant"
          maxLength={100}
          className={inputClass}
        />
        {name.length > 0 && !nameValid && (
          <p className="text-red-400/80 text-xs mt-1 font-mono">
            Name must be 1-100 characters
          </p>
        )}
      </div>

      {/* System prompt */}
      <div>
        <label htmlFor="persona-prompt" className={labelClass}>
          System Prompt
          <span className="text-[#a0a0b8]/30 normal-case tracking-normal ml-2">
            {systemPrompt.length.toLocaleString()} / 50,000
          </span>
        </label>
        <textarea
          id="persona-prompt"
          value={systemPrompt}
          onChange={(e) => setSystemPrompt(e.target.value)}
          placeholder="You are a helpful assistant..."
          rows={6}
          maxLength={50000}
          className={`${inputClass} resize-y min-h-[120px]`}
        />
        {systemPrompt.length > 0 && !promptValid && (
          <p className="text-red-400/80 text-xs mt-1 font-mono">
            System prompt must be 1-50,000 characters
          </p>
        )}
      </div>

      {/* Model */}
      <div>
        <label htmlFor="persona-model" className={labelClass}>
          Model
        </label>
        <input
          id="persona-model"
          type="text"
          value={model}
          onChange={(e) => setModel(e.target.value)}
          placeholder="e.g. gpt-4, claude-3-opus"
          className={inputClass}
        />
        {model.length === 0 && name.length > 0 && (
          <p className="text-red-400/80 text-xs mt-1 font-mono">
            Model is required
          </p>
        )}
      </div>

      {/* Temperature + Max Tokens row */}
      <div className="grid grid-cols-2 gap-4">
        <div>
          <label htmlFor="persona-temp" className={labelClass}>
            Temperature{" "}
            <span className="text-[#e2b714]/60 normal-case tracking-normal">
              {temperature.toFixed(1)}
            </span>
          </label>
          <input
            id="persona-temp"
            type="range"
            min={0}
            max={2}
            step={0.1}
            value={temperature}
            onChange={(e) => setTemperature(Number(e.target.value))}
            className="w-full h-1.5 bg-[#0f3460] rounded-full appearance-none cursor-pointer accent-[#e2b714] [&::-webkit-slider-thumb]:appearance-none [&::-webkit-slider-thumb]:w-4 [&::-webkit-slider-thumb]:h-4 [&::-webkit-slider-thumb]:rounded-full [&::-webkit-slider-thumb]:bg-[#e2b714] [&::-webkit-slider-thumb]:cursor-pointer"
          />
          <div className="flex justify-between text-[#a0a0b8]/30 text-[10px] font-mono mt-1">
            <span>0.0</span>
            <span>1.0</span>
            <span>2.0</span>
          </div>
        </div>

        <div>
          <label htmlFor="persona-max-tokens" className={labelClass}>
            Max Tokens
          </label>
          <input
            id="persona-max-tokens"
            type="number"
            min={1}
            value={maxTokens}
            onChange={(e) => setMaxTokens(Number(e.target.value))}
            className={inputClass}
          />
        </div>
      </div>

      {/* Cooldown + Rate limit row */}
      <div className="grid grid-cols-2 gap-4">
        <div>
          <label htmlFor="persona-cooldown" className={labelClass}>
            Cooldown (seconds)
          </label>
          <input
            id="persona-cooldown"
            type="number"
            min={0}
            value={cooldownSecs}
            onChange={(e) => setCooldownSecs(Number(e.target.value))}
            className={inputClass}
          />
        </div>

        <div>
          <label htmlFor="persona-rate" className={labelClass}>
            Max Tokens / Hour
          </label>
          <input
            id="persona-rate"
            type="number"
            min={0}
            value={maxTokensPerHour}
            onChange={(e) => setMaxTokensPerHour(Number(e.target.value))}
            className={inputClass}
          />
        </div>
      </div>

      {/* Tools */}
      <div>
        <label htmlFor="persona-tools" className={labelClass}>
          Tools Enabled{" "}
          <span className="text-[#a0a0b8]/30 normal-case tracking-normal">
            comma-separated
          </span>
        </label>
        <input
          id="persona-tools"
          type="text"
          value={toolsRaw}
          onChange={(e) => setToolsRaw(e.target.value)}
          placeholder="e.g. web_search, code_exec"
          className={inputClass}
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
              ? "Update Persona"
              : "Create Persona"}
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

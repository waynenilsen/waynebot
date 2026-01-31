import { useErrors } from "../store/ErrorContext";

export default function ErrorBanner() {
  const { errors, dismissError } = useErrors();

  if (errors.length === 0) return null;

  return (
    <div className="fixed top-0 left-0 right-0 z-50 flex flex-col items-center pointer-events-none">
      {errors.map((err) => (
        <div
          key={err.id}
          className="pointer-events-auto mt-2 mx-4 max-w-lg w-full bg-red-900/90 border border-red-500/30 rounded-lg px-4 py-2.5 flex items-start gap-3 shadow-lg backdrop-blur-sm animate-[slideDown_0.2s_ease-out]"
        >
          <span className="text-red-400 text-xs shrink-0 mt-0.5">!</span>
          <p className="text-red-200 text-sm font-mono flex-1 break-words">
            {err.message}
          </p>
          <button
            onClick={() => dismissError(err.id)}
            className="text-red-400/60 hover:text-red-300 text-sm shrink-0 cursor-pointer leading-none"
            aria-label="Dismiss error"
          >
            x
          </button>
        </div>
      ))}
    </div>
  );
}

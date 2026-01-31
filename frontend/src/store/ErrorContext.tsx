import {
  createContext,
  useCallback,
  useContext,
  useMemo,
  useState,
} from "react";
import type { ReactNode } from "react";

export interface AppError {
  id: string;
  message: string;
  dismissedAt?: number;
}

interface ErrorContextValue {
  errors: AppError[];
  pushError: (message: string) => void;
  dismissError: (id: string) => void;
}

const ErrorContext = createContext<ErrorContextValue | null>(null);

let errorSeq = 0;

export function ErrorProvider({ children }: { children: ReactNode }) {
  const [errors, setErrors] = useState<AppError[]>([]);

  const pushError = useCallback((message: string) => {
    const id = `err_${++errorSeq}_${Date.now()}`;
    setErrors((prev) => [...prev.slice(-9), { id, message }]);

    // Auto-dismiss after 8 seconds
    setTimeout(() => {
      setErrors((prev) => prev.filter((e) => e.id !== id));
    }, 8000);
  }, []);

  const dismissError = useCallback((id: string) => {
    setErrors((prev) => prev.filter((e) => e.id !== id));
  }, []);

  const value = useMemo(
    () => ({ errors, pushError, dismissError }),
    [errors, pushError, dismissError],
  );

  return (
    <ErrorContext.Provider value={value}>{children}</ErrorContext.Provider>
  );
}

export function useErrors(): ErrorContextValue {
  const ctx = useContext(ErrorContext);
  if (!ctx) throw new Error("useErrors must be used within ErrorProvider");
  return ctx;
}

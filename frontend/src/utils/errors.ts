/**
 * Extract a human-readable message from an unknown caught error.
 */
export function getErrorMessage(err: unknown): string {
  return err instanceof Error ? err.message : "unknown error";
}

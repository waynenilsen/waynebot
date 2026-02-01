package tools

import "context"

const projectDirKey contextKey = "project_dir"

// WithProjectDir returns a context carrying the given project directory path.
func WithProjectDir(ctx context.Context, dir string) context.Context {
	return context.WithValue(ctx, projectDirKey, dir)
}

// ProjectDirFromContext retrieves the project directory from a context, or "" if not set.
func ProjectDirFromContext(ctx context.Context) string {
	dir, _ := ctx.Value(projectDirKey).(string)
	return dir
}

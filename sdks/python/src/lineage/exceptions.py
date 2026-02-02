"""Exceptions for the Lineage SDK."""


class LineageError(Exception):
    """Base exception for Lineage SDK errors."""

    def __init__(self, message: str, status_code: int | None = None):
        self.message = message
        self.status_code = status_code
        super().__init__(self.message)


class NotFoundError(LineageError):
    """Resource not found (404)."""

    def __init__(self, resource: str, id: str):
        super().__init__(f"{resource} not found: {id}", status_code=404)


class ValidationError(LineageError):
    """Validation error (400)."""

    def __init__(self, message: str):
        super().__init__(message, status_code=400)


class ServerError(LineageError):
    """Server error (5xx)."""

    def __init__(self, message: str = "Internal server error"):
        super().__init__(message, status_code=500)

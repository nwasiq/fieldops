export function ErrorMessage({ message }: { message: string | null }) {
  if (!message) return null;
  return (
    <p className="alert alert-error" role="alert">
      {message}
    </p>
  );
}

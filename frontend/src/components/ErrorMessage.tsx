import { errorMessage, type MessageCode } from "../errors/messages";

interface ErrorMessageProps {
  code: MessageCode;
  position?: number | undefined;
}

export function ErrorMessage({ code, position }: ErrorMessageProps) {
  return (
    <p role="alert" className="error">
      {errorMessage(code, position)}
    </p>
  );
}

import MailFolder from "./MailFolder";

export default function Trash() {
  return <MailFolder folder="Trash" emptyTitle="Trash is empty" emptyDescription="Deleted messages will appear here" />;
}

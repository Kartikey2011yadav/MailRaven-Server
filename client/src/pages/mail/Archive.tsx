import MailFolder from "./MailFolder";

export default function Archive() {
  return <MailFolder folder="Archive" emptyTitle="No archived messages" emptyDescription="Archived messages will appear here" />;
}

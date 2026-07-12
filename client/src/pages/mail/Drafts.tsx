import MailFolder from "./MailFolder";

export default function Drafts() {
  return <MailFolder folder="Drafts" emptyTitle="No drafts" emptyDescription="Saved drafts will appear here" />;
}

"use client";

import SaveTemplateDialog from "@/components/template/save-template-dialog";
import CreateScheduleDialog from "@/components/schedule/create-schedule-dialog";
import QuoteLineCommentsSidebar from "@/components/quote/quote-line-comments-sidebar";

type Props = {
  quoteId?: string;
  projectName: string;
  saveTemplateOpen: boolean;
  onSaveTemplateOpenChange: (open: boolean) => void;
  onSaveQuoteAsTemplate: (name: string) => Promise<boolean>;
  createScheduleOpen: boolean;
  onCreateScheduleOpenChange: (open: boolean) => void;
  commentSidebarOpen: boolean;
  onCommentSidebarOpenChange: (open: boolean) => void;
  commentLineId: string;
  commentLineName: string;
  currentUserId: string;
  currentUserName: string;
};

export default function QuoteFormDialogs({
  quoteId,
  projectName,
  saveTemplateOpen,
  onSaveTemplateOpenChange,
  onSaveQuoteAsTemplate,
  createScheduleOpen,
  onCreateScheduleOpenChange,
  commentSidebarOpen,
  onCommentSidebarOpenChange,
  commentLineId,
  commentLineName,
  currentUserId,
  currentUserName,
}: Props) {
  return (
    <>
      <SaveTemplateDialog
        open={saveTemplateOpen}
        onOpenChange={onSaveTemplateOpenChange}
        defaultName={projectName}
        onSave={onSaveQuoteAsTemplate}
      />

      <CreateScheduleDialog
        open={createScheduleOpen}
        onOpenChange={onCreateScheduleOpenChange}
        initialQuoteId={quoteId}
        lockQuote
      />

      {quoteId && (
        <QuoteLineCommentsSidebar
          open={commentSidebarOpen}
          onOpenChange={onCommentSidebarOpenChange}
          quoteId={quoteId}
          lineId={commentLineId}
          lineName={commentLineName}
          currentUserId={currentUserId}
          currentUserName={currentUserName}
        />
      )}
    </>
  );
}

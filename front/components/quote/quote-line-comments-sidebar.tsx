"use client";

import { useEffect, useRef, useState } from "react";
import { useTranslations } from "next-intl";
import { Loader2Icon, PencilIcon, Trash2Icon } from "lucide-react";
import { toast } from "sonner";
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetFooter,
} from "@/components/ui/sheet";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog";
import {
  listComments,
  createComment,
  updateComment,
  deleteComment,
} from "@/lib/services/comments";
import type { BackendComment } from "@/types/backend";

type Props = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  quoteId: string;
  lineId: string;
  lineName: string;
  currentUserId: string;
  currentUserName: string;
};

function formatDate(iso: string): string {
  try {
    const d = new Date(iso);
    const now = new Date();
    const time = d.toLocaleTimeString("fr-FR", { hour: "2-digit", minute: "2-digit" });
    if (
      d.getDate() === now.getDate() &&
      d.getMonth() === now.getMonth() &&
      d.getFullYear() === now.getFullYear()
    ) {
      return time;
    }
    return (
      d.toLocaleDateString("fr-FR", { day: "2-digit", month: "2-digit", year: "numeric" }) +
      " " +
      time
    );
  } catch {
    return iso;
  }
}

export default function QuoteLineCommentsSidebar({
  open,
  onOpenChange,
  quoteId,
  lineId,
  lineName,
  currentUserId,
  currentUserName,
}: Props) {
  const t = useTranslations("quote.comments");

  const [comments, setComments] = useState<BackendComment[]>([]);
  const [loading, setLoading] = useState(false);
  const [sending, setSending] = useState(false);
  const [newBody, setNewBody] = useState("");
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editBody, setEditBody] = useState("");

  const bottomRef = useRef<HTMLDivElement>(null);
  const prevCommentCountRef = useRef(0);

  useEffect(() => {
    if (!open || !lineId) return;
    setLoading(true);
    listComments(quoteId, lineId).then(({ ok, body }) => {
      setLoading(false);
      if (ok && body.success) {
        setComments((body.comments as BackendComment[]) ?? []);
      } else {
        toast.error(t("loadError"));
      }
    });
  }, [open, quoteId, lineId, t]);

  useEffect(() => {
    if (comments.length > prevCommentCountRef.current) {
      bottomRef.current?.scrollIntoView({ behavior: "smooth" });
    }
    prevCommentCountRef.current = comments.length;
  }, [comments]);

  async function handleSend() {
    const trimmed = newBody.trim();
    if (!trimmed || sending) return;
    setSending(true);
    const { ok, body } = await createComment(quoteId, lineId, trimmed, currentUserName);
    setSending(false);
    if (ok && body.success) {
      setComments((prev) => [...prev, body.comment as BackendComment]);
      setNewBody("");
    } else {
      toast.error((body.message as string) ?? t("loadError"));
    }
  }

  async function handleEditSave(commentId: string) {
    const trimmed = editBody.trim();
    if (!trimmed) return;
    const { ok, body } = await updateComment(quoteId, lineId, commentId, trimmed);
    if (ok && body.success) {
      setComments((prev) =>
        prev.map((c) => (c.comment_id === commentId ? (body.comment as BackendComment) : c)),
      );
      setEditingId(null);
    } else {
      toast.error((body.message as string) ?? t("loadError"));
    }
  }

  async function handleDelete(commentId: string) {
    const { ok, body } = await deleteComment(quoteId, lineId, commentId);
    if (ok && body.success) {
      setComments((prev) => prev.filter((c) => c.comment_id !== commentId));
    } else {
      toast.error((body.message as string) ?? t("loadError"));
    }
  }

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent
        side="right"
        className="flex w-full flex-col sm:max-w-md"
      >
        <SheetHeader>
          <SheetTitle className="truncate pr-8">
            {t("sidebarTitle", { lineName: lineName || "…" })}
          </SheetTitle>
        </SheetHeader>

        <div className="flex-1 overflow-y-auto px-4 py-2">
          {loading && (
            <div className="flex justify-center py-8">
              <Loader2Icon className="text-muted-foreground size-5 animate-spin" />
            </div>
          )}

          {!loading && comments.length === 0 && (
            <p className="text-muted-foreground py-8 text-center text-sm">
              {t("emptyState")}
            </p>
          )}

          {!loading && comments.map((comment) => {
            const isOwn = comment.author_id === currentUserId;
            const isEditing = editingId === comment.comment_id;

            return (
              <div
                key={comment.comment_id}
                className={`mb-4 rounded-lg border p-3 text-sm shadow-sm ${
                  isOwn
                    ? "ml-6 border-primary/20 bg-primary/10"
                    : "mr-6 bg-card"
                }`}
              >
                <p className={`mb-1 font-semibold ${isOwn ? "text-primary" : ""}`}>
                  {comment.author_name}
                </p>

                {isEditing ? (
                  <div className="space-y-2">
                    <Textarea
                      value={editBody}
                      onChange={(e) => setEditBody(e.target.value)}
                      rows={3}
                      autoFocus
                    />
                    <div className="flex gap-2">
                      <Button
                        size="sm"
                        data-testid="comment-save"
                        onClick={() => handleEditSave(comment.comment_id)}
                        disabled={!editBody.trim()}
                      >
                        {t("save")}
                      </Button>
                      <Button
                        size="sm"
                        variant="ghost"
                        data-testid="comment-cancel"
                        onClick={() => setEditingId(null)}
                      >
                        {t("cancel")}
                      </Button>
                    </div>
                  </div>
                ) : (
                  <p className="whitespace-pre-wrap">{comment.body}</p>
                )}

                <div className="mt-2 flex items-center justify-between">
                  <span className="text-muted-foreground text-xs">
                    {formatDate(comment.created_at)}
                  </span>

                  {isOwn && !isEditing && (
                    <div className="flex items-center gap-1">
                      <Button
                        variant="ghost"
                        size="icon-sm"
                        aria-label={t("edit")}
                        onClick={() => {
                          setEditingId(comment.comment_id);
                          setEditBody(comment.body);
                        }}
                      >
                        <PencilIcon className="size-3.5" />
                      </Button>

                      <AlertDialog>
                        <AlertDialogTrigger asChild>
                          <Button
                            variant="ghost"
                            size="icon-sm"
                            aria-label={t("delete")}
                          >
                            <Trash2Icon className="size-3.5" />
                          </Button>
                        </AlertDialogTrigger>
                        <AlertDialogContent>
                          <AlertDialogHeader>
                            <AlertDialogTitle>
                              {t("deleteConfirmTitle")}
                            </AlertDialogTitle>
                            <AlertDialogDescription>
                              {t("deleteConfirmDescription")}
                            </AlertDialogDescription>
                          </AlertDialogHeader>
                          <AlertDialogFooter>
                            <AlertDialogCancel>{t("cancel")}</AlertDialogCancel>
                            <AlertDialogAction
                              onClick={() => handleDelete(comment.comment_id)}
                            >
                              {t("deleteConfirm")}
                            </AlertDialogAction>
                          </AlertDialogFooter>
                        </AlertDialogContent>
                      </AlertDialog>
                    </div>
                  )}
                </div>
              </div>
            );
          })}

          <div ref={bottomRef} />
        </div>

        <SheetFooter className="border-t px-4 pt-3 pb-4">
          <div className="flex w-full flex-col gap-2">
            <Textarea
              placeholder={t("placeholder")}
              value={newBody}
              onChange={(e) => setNewBody(e.target.value)}
              rows={3}
              onKeyDown={(e) => {
                if (e.key === "Enter" && (e.ctrlKey || e.metaKey)) {
                  e.preventDefault();
                  void handleSend();
                }
              }}
            />
            <Button
              onClick={handleSend}
              disabled={!newBody.trim() || sending}
              className="self-end"
            >
              {sending ? (
                <Loader2Icon className="size-4 animate-spin" />
              ) : null}
              {t("send")}
            </Button>
          </div>
        </SheetFooter>
      </SheetContent>
    </Sheet>
  );
}

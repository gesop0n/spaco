import { useState } from "react";
import type { FormEvent } from "react";
import { useReviewWorkspace } from "../ReviewWorkspace";
import { catalog } from "../ReviewWorkspace/fixtures";
import { addDays, problemFromUrl } from "../ReviewWorkspace/model";
import type { Problem } from "../ReviewWorkspace/types";

const contests = ["abc350", "abc351", "abc352"];

export function useProblemRegistrationPage() {
  const { problems, dispatch, today } = useReviewWorkspace();
  const [mode, setMode] = useState<"contest" | "url">("contest");
  const [query, setQuery] = useState("");
  const [contestId, setContestId] = useState("abc350");
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [note, setNote] = useState("");
  const [url, setUrl] = useState("");
  const [urlError, setUrlError] = useState("");
  const [notice, setNotice] = useState<{ count: number; firstId: string }>();
  const registered = new Set(problems.map((problem) => problem.id));
  const normalizedQuery = query.toLowerCase().replace(/\s/g, "");
  const visibleContests = contests.filter((id) =>
    `${id}atcoderbeginnercontest${id.slice(3)}`.includes(normalizedQuery),
  );
  const activeContest = visibleContests.includes(contestId) ? contestId : visibleContests[0];
  const contestProblems = catalog.filter((problem) => problem.contestId === activeContest);
  const available = contestProblems.filter((problem) => !registered.has(problem.id));
  const selectedProblems = catalog.filter(
    (problem) => selected.has(problem.id) && !registered.has(problem.id),
  );

  function registerProblems(additions: Problem[]) {
    if (!additions.length) return;
    dispatch({ type: "register", problems: additions, today, note });
    setNotice({ count: additions.length, firstId: additions[0].id });
    setSelected(new Set());
    setNote("");
  }

  function toggleProblem(id: string) {
    setSelected((previous) => {
      const next = new Set(previous);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  }

  function toggleAll() {
    setSelected((previous) => {
      const next = new Set(previous);
      const allSelected = available.every((problem) => previous.has(problem.id));
      for (const problem of available) {
        if (allSelected) next.delete(problem.id);
        else next.add(problem.id);
      }
      return next;
    });
  }

  function submitUrl(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setUrlError("");
    const problem = problemFromUrl(url, catalog);
    if (!problem) {
      setUrlError("https://atcoder.jp/contests/…/tasks/… の問題URLを入力してください。");
      return;
    }
    if (registered.has(problem.id)) {
      setUrlError("この問題は登録済みです。復習リストから確認できます。");
      return;
    }
    registerProblems([problem]);
    setUrl("");
  }

  return {
    mode,
    setMode,
    query,
    setQuery,
    setContestId,
    activeContest,
    visibleContests,
    contestProblems,
    registered,
    selected,
    selectedCount: selectedProblems.length,
    allSelected: available.length > 0 && available.every((problem) => selected.has(problem.id)),
    hasAvailable: available.length > 0,
    toggleProblem,
    toggleAll,
    note,
    setNote,
    url,
    setUrl,
    urlError,
    notice,
    nextDay: addDays(today, 1),
    submitUrl,
    registerSelected: () => registerProblems(selectedProblems),
  };
}

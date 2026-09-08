import { addDays } from "./model";
import type { Problem, StudyProblem } from "./types";

// 問題名のみ公式問題一覧を参照。学習状況・メモ・日程はすべて架空のサンプル。
export const catalog: Problem[] = [
  { id: "abc350_a", contestId: "abc350", index: "A", title: "Past ABCs" },
  { id: "abc350_b", contestId: "abc350", index: "B", title: "Dentist Aoki" },
  { id: "abc350_c", contestId: "abc350", index: "C", title: "Sort" },
  { id: "abc350_d", contestId: "abc350", index: "D", title: "New Friends" },
  { id: "abc351_a", contestId: "abc351", index: "A", title: "The bottom of the ninth" },
  { id: "abc351_b", contestId: "abc351", index: "B", title: "Spot the Difference" },
  { id: "abc351_c", contestId: "abc351", index: "C", title: "Merge the balls" },
  { id: "abc351_d", contestId: "abc351", index: "D", title: "Grid and Magnet" },
  { id: "abc352_a", contestId: "abc352", index: "A", title: "AtCoder Line" },
  { id: "abc352_b", contestId: "abc352", index: "B", title: "Typing" },
  { id: "abc352_c", contestId: "abc352", index: "C", title: "Standing On The Shoulders" },
  { id: "abc352_d", contestId: "abc352", index: "D", title: "Permutation Subsequence" },
];

export function sampleProblems(today: string): StudyProblem[] {
  return [
    { id: "abc350_d", offset: -2, note: "連結成分ごとに考える発想を、自分で思い出したい。" },
    { id: "abc351_c", offset: 0, note: "同じ値が続くときの処理を、もう一度整理する。" },
    { id: "abc352_c", offset: 0, note: "何を最後に選ぶとよいか、式を書いて考える。" },
    { id: "abc350_c", offset: 1, note: "位置を管理するときの更新を忘れない。" },
    { id: "abc351_d", offset: 3, note: "移動できるマスと境界のマスを分けてみる。" },
  ].map(({ id, offset, note }, index) => ({
    ...catalog.find((problem) => problem.id === id)!,
    registeredOn: addDays(today, -10),
    dueOn: addDays(today, offset),
    paused: false,
    registrationNote: note,
    history:
      index === 2
        ? []
        : [
            {
              id: `sample-${id}`,
              result: index % 2 === 0 ? "retry" : "assisted",
              performedAt: `${addDays(today, -4)}T20:30`,
              note: "解法の方針を確認した。次は何も見ずに取り組みたい。",
            },
          ],
  }));
}

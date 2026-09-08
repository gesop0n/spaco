export const defaultTimeZone = "Asia/Tokyo";

type TimeZoneOption = { value: string; label: string };

const commonTimeZones: TimeZoneOption[] = [
  { value: defaultTimeZone, label: "日本（Asia/Tokyo）" },
  { value: "UTC", label: "協定世界時（UTC）" },
  { value: "Asia/Seoul", label: "ソウル（Asia/Seoul）" },
  { value: "Asia/Shanghai", label: "上海（Asia/Shanghai）" },
  { value: "Asia/Taipei", label: "台北（Asia/Taipei）" },
  { value: "Asia/Singapore", label: "シンガポール（Asia/Singapore）" },
  { value: "Asia/Kolkata", label: "インド（Asia/Kolkata）" },
  { value: "Europe/London", label: "ロンドン（Europe/London）" },
  { value: "Europe/Paris", label: "パリ（Europe/Paris）" },
  { value: "America/New_York", label: "ニューヨーク（America/New_York）" },
  { value: "America/Chicago", label: "シカゴ（America/Chicago）" },
  { value: "America/Los_Angeles", label: "ロサンゼルス（America/Los_Angeles）" },
  { value: "Pacific/Honolulu", label: "ハワイ（Pacific/Honolulu）" },
  { value: "Australia/Sydney", label: "シドニー（Australia/Sydney）" },
  { value: "Pacific/Auckland", label: "オークランド（Pacific/Auckland）" },
];

const supportedTimeZones =
  typeof Intl.supportedValuesOf === "function" ? Intl.supportedValuesOf("timeZone") : [];

/** 一覧に含まれない別名なども、保存済みの値を失わずに選択・再保存できるようにする。 */
export function getTimeZoneGroups(currentTimeZone: string = defaultTimeZone) {
  const commonValues = new Set(commonTimeZones.map(({ value }) => value));
  const otherValues = new Set(supportedTimeZones);
  if (currentTimeZone) otherValues.add(currentTimeZone);

  return [
    { label: "主なタイムゾーン", options: commonTimeZones },
    {
      label: "その他のタイムゾーン",
      options: [...otherValues]
        .filter((value) => !commonValues.has(value))
        .sort()
        .map((value) => ({ value, label: value })),
    },
  ].filter(({ options }) => options.length > 0);
}

import i18n from "i18next";
import { initReactI18next } from "react-i18next";
import en from "./locales/en.json";
import zh from "./locales/zh.json";

function getSavedLanguage() {
  try {
    const saved = JSON.parse(localStorage.getItem("language") ?? "");
    if (saved === "zh" || saved === "en") {
      return saved;
    }
  } catch {
    // 忽略损坏的值
  }
  const browserLang = navigator.language.toLowerCase();
  return browserLang.startsWith("zh") ? "zh" : "en";
}

i18n.use(initReactI18next).init({
  resources: {
    en: { translation: en },
    zh: { translation: zh },
  },
  lng: getSavedLanguage(),
  fallbackLng: "zh",
  interpolation: {
    escapeValue: false,
  },
});

import { languageAtom, themeAtom, type Theme } from "@/stores/settings";
import { useAtom } from "jotai";
import { Sun, Moon, Monitor, Languages, Check } from "lucide-react";
import { useTranslation } from "react-i18next";

import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

function SettingsBar() {
  const [language, setLanguage] = useAtom(languageAtom);
  const [theme, setTheme] = useAtom(themeAtom);
  const { t, i18n } = useTranslation();

  const handleLanguageChange = (newLang: "zh" | "en") => {
    setLanguage(newLang);
    i18n.changeLanguage(newLang);
  };

  const getThemeIcon = () => {
    switch (theme) {
      case "light":
        return <Sun />;
      case "dark":
        return <Moon />;
      default:
        return <Monitor />;
    }
  };

  const themeOptions: { value: Theme; icon: typeof Sun; label: string }[] = [
    { value: "light", icon: Sun, label: t("theme.light") },
    { value: "dark", icon: Moon, label: t("theme.dark") },
    { value: "system", icon: Monitor, label: t("theme.system") },
  ];

  return (
    <div className="flex items-center gap-1">
      <DropdownMenu>
        <DropdownMenuTrigger
          render={<Button variant="ghost" size="sm" className="rounded-full" />}
        >
          <Languages data-icon="inline-start" />
          <span>{language === "zh" ? "中文" : "EN"}</span>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          <DropdownMenuGroup>
            {(["zh", "en"] as const).map((lang) => (
              <DropdownMenuItem
                key={lang}
                onClick={() => handleLanguageChange(lang)}
              >
                {lang === "zh" ? "中文" : "English"}
                {language === lang && (
                  <Check data-icon="inline-end" className="text-primary" />
                )}
              </DropdownMenuItem>
            ))}
          </DropdownMenuGroup>
        </DropdownMenuContent>
      </DropdownMenu>

      <DropdownMenu>
        <DropdownMenuTrigger
          render={
            <Button
              variant="ghost"
              size="icon"
              className="size-9 rounded-full"
              aria-label={t("theme.system")}
            />
          }
        >
          {getThemeIcon()}
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          <DropdownMenuGroup>
            {themeOptions.map(({ value, icon: Icon, label }) => (
              <DropdownMenuItem key={value} onClick={() => setTheme(value)}>
                <Icon data-icon="inline-start" />
                {label}
                {theme === value && (
                  <Check data-icon="inline-end" className="text-primary" />
                )}
              </DropdownMenuItem>
            ))}
          </DropdownMenuGroup>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  );
}

export default SettingsBar;

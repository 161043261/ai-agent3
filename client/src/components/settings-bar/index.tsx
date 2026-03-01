import { languageAtom, themeAtom, type Theme } from "@/stores/settings";
import { useAtom } from "jotai";
import { Sun, Moon, Monitor, Languages, Check } from "lucide-react";
import { useTranslation } from "react-i18next";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "../ui/dropdown-menu";
import { Button } from "../ui/button";

function SettingsBar() {
  const { t, i18n } = useTranslation();
  const [theme, setTheme] = useAtom(themeAtom);
  const [language, setLanguage] = useAtom(languageAtom);

  const handleLanguageChange = (lang: "zh" | "en") => {
    setLanguage(lang);
    i18n.changeLanguage(lang);
  };

  const handleThemeChange = (newTheme: Theme) => {
    setTheme(newTheme);
  };

  const getThemeIcon = () => {
    switch (theme) {
      case "light":
        return <Sun className="h-4 w-4" />;
      case "dark":
        return <Moon className="h-4 w-4" />;
      default:
        return <Monitor className="h-4 w-4" />;
    }
  };

  return (
    <div className="flex items-center gap-1">
      {/* Language Switcher */}
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button
            variant="ghost"
            size="sm"
            className="text-muted-foreground hover:text-foreground hover:bg-muted h-9 gap-2 rounded-full px-3"
          >
            <Languages className="h-4 w-4" />
            <span className="text-sm">{language === "zh" ? "中文" : "EN"}</span>
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end" className="min-w-30">
          <DropdownMenuItem
            onClick={() => handleLanguageChange("zh")}
            className="flex cursor-pointer items-center justify-between"
          >
            <span>中文</span>
            {language === "zh" && <Check className="text-primary h-4 w-4" />}
          </DropdownMenuItem>
          <DropdownMenuItem
            onClick={() => handleLanguageChange("en")}
            className="flex cursor-pointer items-center justify-between"
          >
            <span>English</span>
            {language === "en" && <Check className="text-primary h-4 w-4" />}
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>

      {/* Theme Switcher */}
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button
            variant="ghost"
            size="sm"
            className="text-muted-foreground hover:text-foreground hover:bg-muted h-9 w-9 rounded-full p-0"
          >
            {getThemeIcon()}
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end" className="min-w-35">
          <DropdownMenuItem
            onClick={() => handleThemeChange("light")}
            className="flex cursor-pointer items-center justify-between"
          >
            <div className="flex items-center gap-2">
              <Sun className="h-4 w-4" />
              <span>{t("theme.light")}</span>
            </div>
            {theme === "light" && <Check className="text-primary h-4 w-4" />}
          </DropdownMenuItem>
          <DropdownMenuItem
            onClick={() => handleThemeChange("dark")}
            className="flex cursor-pointer items-center justify-between"
          >
            <div className="flex items-center gap-2">
              <Moon className="h-4 w-4" />
              <span>{t("theme.dark")}</span>
            </div>
            {theme === "dark" && <Check className="text-primary h-4 w-4" />}
          </DropdownMenuItem>
          <DropdownMenuItem
            onClick={() => handleThemeChange("system")}
            className="flex cursor-pointer items-center justify-between"
          >
            <div className="flex items-center gap-2">
              <Monitor className="h-4 w-4" />
              <span>{t("theme.system")}</span>
            </div>
            {theme === "system" && <Check className="text-primary h-4 w-4" />}
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  );
}

export default SettingsBar;

import { useLogin } from "@/hooks/queries";
import SettingsBar from "@/components/settings-bar";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { setTokenAtom } from "@/stores/auth";
import { zodResolver } from "@hookform/resolvers/zod";
import { useSetAtom } from "jotai";
import { MessageSquare } from "lucide-react";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { Link, useNavigate } from "react-router-dom";
import { toast } from "sonner";
import z from "zod";

const loginSchema = z.object({
  username: z.string().min(1, "auth.username_required"),
  password: z.string().min(6, "auth.password_required"),
});

type LoginFormData = z.infer<typeof loginSchema>;

function Login() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const setToken = useSetAtom(setTokenAtom);
  const loginMutation = useLogin();
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginFormData>({
    resolver: zodResolver(loginSchema),
    defaultValues: {
      username: "",
      password: "",
    },
  });

  const onSubmit = (data: LoginFormData) => {
    loginMutation.mutate(
      { username: data.username, password: data.password },
      {
        onSuccess: ({ code, token, message }) => {
          if (code === 1000 && token) {
            setToken(token);
            toast.success(t("auth.login_success"));
            navigate("/menu");
          } else {
            toast.error(message || t("auth.login_failed"));
          }
        },
        onError: (err) => {
          if (import.meta.env.DEV) console.error(err);
          toast.error(t("auth.login_failed"));
        },
      },
    );
  };

  return (
    <div className="bg-background flex min-h-screen flex-col">
      {/* Settings Bar - Top Right */}
      <div className="absolute top-4 right-4 z-10">
        <SettingsBar />
      </div>

      <div className="flex flex-1 items-center justify-center">
        <Card className="border-border bg-card w-105 rounded-lg border shadow-none">
          <CardHeader className="pt-10 pb-2 text-center">
            {/* Logo */}
            <div className="mb-6 flex justify-center">
              <div className="bg-accent flex h-16 w-16 items-center justify-center rounded-full">
                <MessageSquare className="text-primary h-8 w-8" />
              </div>
            </div>
            <CardTitle className="text-foreground text-2xl font-normal">
              {t("auth.login")}
            </CardTitle>
            <p className="text-muted-foreground mt-2 text-sm">
              {t("auth.continue_with")}
            </p>
          </CardHeader>
          <CardContent className="px-10 pb-10">
            <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
              <div className="space-y-2">
                <Label
                  htmlFor="username"
                  className="text-muted-foreground text-xs font-medium"
                >
                  {t("auth.username")}
                </Label>
                <Input
                  id="username"
                  placeholder={t("auth.username_required")}
                  {...register("username")}
                  className="border-input focus:border-primary focus:ring-primary bg-background h-12 rounded-md transition-colors"
                />
                {errors.username && (
                  <p className="text-destructive text-sm">
                    {t(errors.username.message || "")}
                  </p>
                )}
              </div>

              <div className="space-y-2">
                <Label
                  htmlFor="password"
                  className="text-muted-foreground text-xs font-medium"
                >
                  {t("auth.password")}
                </Label>
                <Input
                  id="password"
                  type="password"
                  placeholder={t("auth.password_required")}
                  {...register("password")}
                  className="border-input focus:border-primary focus:ring-primary bg-background h-12 rounded-md transition-colors"
                />
                {errors.password && (
                  <p className="text-destructive text-sm">
                    {t(errors.password.message || "")}
                  </p>
                )}
              </div>

              <div className="flex items-center justify-between pt-4">
                <Link
                  to="/register"
                  className="text-primary hover:text-primary/80 text-sm font-medium"
                >
                  {t("auth.no_account")}
                </Link>
                <Button
                  type="submit"
                  className="bg-primary hover:bg-primary/90 text-primary-foreground h-10 rounded-md px-6 font-medium transition-colors"
                  disabled={loginMutation.isPending}
                >
                  {loginMutation.isPending
                    ? t("auth.signing_in")
                    : t("common.next")}
                </Button>
              </div>
            </form>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

export default Login;

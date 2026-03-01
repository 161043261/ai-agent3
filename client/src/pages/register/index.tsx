import { useTranslation } from "react-i18next";
import { z } from "zod";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Link, useNavigate } from "react-router-dom";
import { useSetAtom } from "jotai";
import { setTokenAtom } from "@/stores/auth";
import { useRegister } from "@/hooks/queries";
import { toast } from "sonner";
import SettingsBar from "@/components/settings-bar";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { UserPlus } from "lucide-react";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";

const registerSchema = z
  .object({
    email: z.string().email("auth.email_invalid"),
    password: z.string().min(6, "auth.password_required"),
    confirmPassword: z.string().min(1, "auth.confirm_password_required"),
  })
  // refine 自定义校验逻辑
  .refine((data) => data.password === data.confirmPassword, {
    message: "auth.password_mismatch",
    // path 绑定到 confirmPassword 字段
    path: ["confirmPassword"],
  });

type RegisterFormData = z.infer<typeof registerSchema>;

function Register() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const setToken = useSetAtom(setTokenAtom);
  const registerMutation = useRegister();

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<RegisterFormData>({
    resolver: zodResolver(registerSchema),
    defaultValues: {
      email: "",
      password: "",
      confirmPassword: "",
    },
  });

  const onSubmit = (data: RegisterFormData) => {
    registerMutation.mutate(
      { email: data.email, password: data.password },
      {
        onSuccess: ({ code, token, message }) => {
          if (code === 1000 && token) {
            setToken(token);
            toast.success(t("auth.register_success"));
            navigate("/menu");
          } else {
            toast.error(message || t("auth.register_failed"));
          }
        },
        onError: (err) => {
          if (import.meta.env.DEV) console.error(err);
          toast.error(t("auth.register_failed"));
        },
      },
    );
  };

  return (
    <div className="bg-background flex min-h-screen flex-col">
      <div className="absolute top-4 right-4 z-10">
        <SettingsBar />
      </div>
      <div className="flex flex-1 items-center justify-center">
        <Card className="border-border bg-card w-105 rounded-lg border shadow-none">
          <CardHeader className="pt-10 pb-2 text-center">
            {/* Logo */}
            <div className="mb-6 flex justify-center">
              <div className="bg-accent flex h-16 w-16 items-center justify-center rounded-full">
                <UserPlus className="text-primary h-8 w-8" />
              </div>
            </div>
            <CardTitle className="text-foreground text-2xl font-normal">
              {t("auth.register")}
            </CardTitle>
            <p className="text-muted-foreground mt-2 text-sm">
              {t("auth.register_for")}
            </p>
          </CardHeader>
          <CardContent className="px-10 pb-10">
            <form onSubmit={handleSubmit(onSubmit)} className="space-y-5">
              <div className="space-y-2">
                <Label
                  htmlFor="email"
                  className="text-muted-foreground text-xs font-medium"
                >
                  {t("auth.email")}
                </Label>
                <Input
                  id="email"
                  type="email"
                  placeholder={t("auth.email_required")}
                  {...register("email")}
                  className="border-input focus:border-primary focus:ring-primary bg-background h-12 rounded-md"
                />
                {errors.email && (
                  <p className="text-destructive text-sm">
                    {t(errors.email.message ?? "")}
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
                  className="border-input focus:border-primary focus:ring-primary bg-background h-12 rounded-md"
                />
                {errors.password && (
                  <p className="text-destructive text-sm">
                    {t(errors.password.message ?? "")}
                  </p>
                )}
              </div>

              <div className="space-y-2">
                <Label
                  htmlFor="confirmPassword"
                  className="text-muted-foreground text-xs font-medium"
                >
                  {t("auth.confirm_password")}
                </Label>
                <Input
                  id="confirmPassword"
                  type="password"
                  placeholder={t("auth.password_mismatch")}
                  {...register("confirmPassword")}
                  className="border-input focus:border-primary focus:ring-primary bg-background h-12 rounded-md"
                />
                {errors.confirmPassword && (
                  <p className="text-destructive text-sm">
                    {t(errors.confirmPassword.message ?? "")}
                  </p>
                )}
              </div>

              <div className="flex items-center justify-between pt-4">
                <Link
                  to="/login"
                  className="text-primary hover:text-primary/80 text-sm font-medium"
                >
                  {t("auth.has_account")}
                </Link>
                <Button
                  type="submit"
                  className="bg-primary hover:bg-primary/90 text-primary-foreground h-10 rounded-md px-6 font-medium"
                  disabled={registerMutation.isPending}
                >
                  {registerMutation.isPending
                    ? t("auth.signing_up")
                    : t("auth.register")}
                </Button>
              </div>
            </form>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

export default Register;

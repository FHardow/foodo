import type { KcContext } from "../KcContext";
import type { I18n } from "../i18n";
import Template from "../Template";

type LoginKcContext = Extract<KcContext, { pageId: "login.ftl" }>;

type Props = {
    kcContext: LoginKcContext;
    i18n: I18n;
};

export default function Login({ kcContext }: Props) {
    const { url, realm, messagesPerField, message } = kcContext;

    return (
        <Template>
            <h1 className="text-lg font-bold text-[#3d2b1a] mb-6">Sign in</h1>

            {message && message.type === "error" && (
                <p className="text-sm text-red-600 mb-4">{message.summary}</p>
            )}

            <form action={url.loginAction} method="post" className="space-y-4">
                <div>
                    <label
                        htmlFor="username"
                        className="block text-sm font-semibold text-[#5c3d1e] mb-1"
                    >
                        Email
                    </label>
                    <input
                        id="username"
                        name="username"
                        type="email"
                        autoComplete="email"
                        className="w-full bg-[#faf7f2] border border-[#e8ddd0] rounded-lg px-3 py-2 text-[#3d2b1a] text-sm focus:outline-none focus:ring-2 focus:ring-[#5c3d1e] focus:border-transparent"
                    />
                    {messagesPerField.existsError("username") && (
                        <p className="text-xs text-red-600 mt-1">
                            {messagesPerField.getFirstError("username")}
                        </p>
                    )}
                </div>

                <div>
                    <label
                        htmlFor="password"
                        className="block text-sm font-semibold text-[#5c3d1e] mb-1"
                    >
                        Password
                    </label>
                    <input
                        id="password"
                        name="password"
                        type="password"
                        autoComplete="current-password"
                        className="w-full bg-[#faf7f2] border border-[#e8ddd0] rounded-lg px-3 py-2 text-[#3d2b1a] text-sm focus:outline-none focus:ring-2 focus:ring-[#5c3d1e] focus:border-transparent"
                    />
                    {messagesPerField.existsError("password") && (
                        <p className="text-xs text-red-600 mt-1">
                            {messagesPerField.getFirstError("password")}
                        </p>
                    )}
                </div>

                <button
                    type="submit"
                    className="w-full bg-[#5c3d1e] text-white rounded-lg py-2.5 text-sm font-semibold hover:bg-[#3d2b1a] transition-colors mt-2"
                >
                    Sign in
                </button>
            </form>

            <div className="flex justify-between mt-4">
                {realm.resetPasswordAllowed && (
                    <a
                        href={url.loginResetCredentialsUrl}
                        className="text-sm text-[#5c3d1e] hover:underline"
                    >
                        Forgot password?
                    </a>
                )}
                {realm.registrationAllowed && (
                    <a
                        href={url.registrationUrl}
                        className="text-sm text-[#5c3d1e] hover:underline"
                    >
                        Create account
                    </a>
                )}
            </div>
        </Template>
    );
}

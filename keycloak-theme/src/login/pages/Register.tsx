import type { KcContext } from "../KcContext";
import type { I18n } from "../i18n";
import Template from "../Template";

type RegisterKcContext = Extract<KcContext, { pageId: "register.ftl" }>;

type Props = {
    kcContext: RegisterKcContext;
    i18n: I18n;
};

const inputClass =
    "w-full bg-[#faf7f2] border border-[#e8ddd0] rounded-lg px-3 py-2 text-[#3d2b1a] text-sm focus:outline-none focus:ring-2 focus:ring-[#5c3d1e] focus:border-transparent";

const labelClass = "block text-sm font-semibold text-[#5c3d1e] mb-1";

export default function Register({ kcContext }: Props) {
    const { url, messagesPerField, message } = kcContext;

    return (
        <Template>
            <h1 className="text-lg font-bold text-[#3d2b1a] mb-6">Create account</h1>

            {message && message.type === "error" && (
                <p className="text-sm text-red-600 mb-4">{message.summary}</p>
            )}

            <form action={url.registrationAction} method="post" className="space-y-4">
                <div className="flex gap-3">
                    <div className="flex-1">
                        <label htmlFor="firstName" className={labelClass}>
                            First name
                        </label>
                        <input
                            id="firstName"
                            name="firstName"
                            type="text"
                            autoComplete="given-name"
                            className={inputClass}
                        />
                        {messagesPerField.existsError("firstName") && (
                            <p className="text-xs text-red-600 mt-1">
                                {messagesPerField.getFirstError("firstName")}
                            </p>
                        )}
                    </div>
                    <div className="flex-1">
                        <label htmlFor="lastName" className={labelClass}>
                            Last name
                        </label>
                        <input
                            id="lastName"
                            name="lastName"
                            type="text"
                            autoComplete="family-name"
                            className={inputClass}
                        />
                        {messagesPerField.existsError("lastName") && (
                            <p className="text-xs text-red-600 mt-1">
                                {messagesPerField.getFirstError("lastName")}
                            </p>
                        )}
                    </div>
                </div>

                <div>
                    <label htmlFor="email" className={labelClass}>
                        Email
                    </label>
                    <input
                        id="email"
                        name="email"
                        type="email"
                        autoComplete="email"
                        className={inputClass}
                    />
                    {messagesPerField.existsError("email") && (
                        <p className="text-xs text-red-600 mt-1">
                            {messagesPerField.getFirstError("email")}
                        </p>
                    )}
                </div>

                <div>
                    <label htmlFor="password" className={labelClass}>
                        Password
                    </label>
                    <input
                        id="password"
                        name="password"
                        type="password"
                        autoComplete="new-password"
                        className={inputClass}
                    />
                    {messagesPerField.existsError("password") && (
                        <p className="text-xs text-red-600 mt-1">
                            {messagesPerField.getFirstError("password")}
                        </p>
                    )}
                </div>

                <div>
                    <label htmlFor="password-confirm" className={labelClass}>
                        Confirm password
                    </label>
                    <input
                        id="password-confirm"
                        name="password-confirm"
                        type="password"
                        autoComplete="new-password"
                        className={inputClass}
                    />
                    {messagesPerField.existsError("password-confirm") && (
                        <p className="text-xs text-red-600 mt-1">
                            {messagesPerField.getFirstError("password-confirm")}
                        </p>
                    )}
                </div>

                <button
                    type="submit"
                    className="w-full bg-[#5c3d1e] text-white rounded-lg py-2.5 text-sm font-semibold hover:bg-[#3d2b1a] transition-colors mt-2"
                >
                    Create account
                </button>
            </form>

            <hr className="border-[#e8ddd0] my-5" />

            <p className="text-center text-sm">
                <a href={url.loginUrl} className="text-[#5c3d1e] hover:underline">
                    Already have an account? Sign in
                </a>
            </p>
        </Template>
    );
}

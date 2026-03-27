import type { ReactNode } from "react";

type Props = {
    children: ReactNode;
};

export default function Template({ children }: Props) {
    return (
        <div className="min-h-screen bg-[#faf7f2] flex items-center justify-center px-4 py-12">
            <div className="w-full max-w-md bg-white border border-[#e8ddd0] rounded-lg shadow-sm p-8">
                <div className="text-center mb-6">
                    <p className="text-xl font-bold text-[#5c3d1e]">🍞 Bread Order</p>
                    <p className="text-sm text-[#8a6a50] mt-1">Your local bakery</p>
                </div>
                {children}
            </div>
        </div>
    );
}

type StatCardProps = {
    label: string;
    value: number | string;
    color?: "blue" | "green" | "purple" | "yellow" | "red";
    subtitle?: string;
};

export function StatCard({ label, value, color = "blue", subtitle }: StatCardProps) {
    const colorClasses = {
        blue: "bg-blue-500/10 border-blue-500/30 text-blue-400",
        green: "bg-green-500/10 border-green-500/30 text-green-400",
        purple: "bg-purple-500/10 border-purple-500/30 text-purple-400",
        yellow: "bg-yellow-500/10 border-yellow-500/30 text-yellow-400",
        red: "bg-red-500/10 border-red-500/30 text-red-400",
    }[color];

    return (
        <div className={`${colorClasses} rounded-lg border-2 p-4 shadow-sm`}>
            <div className='text-center'>
                <div className='mb-1 font-bold text-2xl'>{value}</div>
                <div className='font-semibold text-xs uppercase tracking-wide'>{label}</div>
                {subtitle && <div className='mt-1 text-xs opacity-75'>{subtitle}</div>}
            </div>
        </div>
    );
}

type EmptyStateProps = {
    title: string;
    description: string;
    icon?: string;
};

export function EmptyState({ title, description, icon = "🔍" }: EmptyStateProps) {
    return (
        <div className='flex flex-col items-center justify-center rounded-xl border-2 border-border-primary bg-bg-secondary px-6 py-16'>
            <div className='mb-4 text-6xl'>{icon}</div>
            <h3 className='mb-2 font-bold text-text-primary text-xl'>{title}</h3>
            <p className='max-w-md text-center text-sm text-text-secondary'>{description}</p>
        </div>
    );
}

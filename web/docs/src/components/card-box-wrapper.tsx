type Props = {
    title: string;
    children?: React.ReactNode;
};
export const CardBoxWrapper = ({ title, children }: Props) => {
    return (
        <div className='mb-6 rounded-xl border-2 border-border-primary bg-bg-secondary p-6 shadow-md transition-colors duration-200 hover:border-accent-blue'>
            <h2 className='mb-4 font-bold text-text-primary text-xl'>{title}</h2>
            {children}
        </div>
    );
};

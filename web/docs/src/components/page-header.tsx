type PageHeaderProps = {
    title: string;
    description: string;
};

export const PageHeader = ({ title, description }: PageHeaderProps) => {
    return (
        <div className='mb-8'>
            <h1 className='mb-3 font-bold text-4xl text-text-primary'>{title}</h1>
            <p className='text-lg text-text-secondary'>{description}</p>
        </div>
    );
};

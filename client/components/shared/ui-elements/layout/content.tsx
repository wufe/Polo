import React from 'react';

export const Content = ({ children }: React.PropsWithChildren<{}>) =>
    <div className="font-quicksand w-full py-8 pb-12">
        <div className="w-full mx-auto lg:max-w-1500 px-5">
            {children}
        </div>
    </div>;
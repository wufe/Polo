import React, { useEffect, useState } from 'react';
import { observer } from 'mobx-react-lite';
import { IApp, IApplication } from '@/state/models';
import { Application } from './application/application';
import { values } from 'mobx';

type TProps = {
    app: IApp;
}

export const Dashboard = observer((props: TProps) => {

    const [openApplications, setOpenApplications] = useState<{[k:string]: boolean}>({});
    const [openToggleEnabled, setOpenToggleEnabled] = useState(false);

    const requestData = async () => {
        await props.app.retrieveApplications();
        await props.app.retrieveAllSessions();
    }

    useEffect(() => {
        requestData();
        const interval = setInterval(() => requestData(), 2000);
        return () => clearInterval(interval);
    }, []);

    useEffect(() => {
        setOpenToggleEnabled(props.app.applications.size > 1);
    }, [props.app.applications.size])

    const toggleApplication = (name: string) => () => {
        if (props.app.applications.size > 1)
            setOpenApplications(a => ({ ...a, [name]: !a[name] }));
    }

    return <div className="font-quicksand w-full py-8 pb-12">
        <div className="w-10/12 mx-auto">
            <h1 className="text-4xl mb-10 font-light text-nord1 dark:text-nord5">Applications</h1>
            {(values(props.app.applications) as any as IApplication[]).map((application, index) =>
                <Application
                    isOpen={!!openApplications[application.name] || !openToggleEnabled}
                    onToggle={toggleApplication(application.name)}
                    toggleEnabled={openToggleEnabled}
                    key={index}
                    sessions={props.app.sessionsByApplicationName[application.name]}
                    application={application} />)}
        </div>
        
    </div>;
})

export default Dashboard;
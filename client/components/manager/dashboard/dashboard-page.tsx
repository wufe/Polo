import React, { useEffect } from 'react';
import { observer } from 'mobx-react-lite';
import { IApp, IApplication } from '@/state/models';
import { Application } from './application/application';
import { values } from 'mobx';

type TProps = {
    app: IApp;
}

export const Dashboard = observer((props: TProps) => {

    const requestData = async () => {
        await props.app.retrieveApplications();
        await props.app.retrieveAllSessions();
    }

    useEffect(() => {

        requestData();

        const interval = setInterval(() => requestData(), 2000);
        
        return () => clearInterval(interval);
    }, [])

    return <div className="font-quicksand w-10/12 p-20 mx-auto">
        <h1 className="text-4xl mb-3 text-nord1 dark:text-nord5">Applications</h1>
        {(values(props.app.applications) as any as IApplication[]).map((application, index) =>
            <Application
                key={index}
                sessions={props.app.sessionsByApplicationName[application.name]}
                application={application} />)}
    </div>;
})

export default Dashboard;
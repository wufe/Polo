import React, { useEffect } from 'react';
import { observer } from 'mobx-react-lite';
import { IApp, IService } from '@/state/models';
import { Service } from './service/service';
import './dashboard.scss';
import { values } from 'mobx';

type TProps = {
    app: IApp;
}

export const Dashboard = observer((props: TProps) => {

    const requestData = async () => {
        await props.app.retrieveServices();
        await props.app.retrieveAllSessions();
    }

    useEffect(() => {

        requestData();

        const interval = setInterval(() => requestData(), 2000);
        
        return () => clearInterval(interval);
    }, [])

    return <div className="dashboard__component font-quicksand w-10/12">
        <h1 className="text-4xl mb-3 text-nord1 dark:text-nord5">Services</h1>
        {(values(props.app.services) as any as IService[]).map((service, index) =>
            <Service
                key={index}
                sessions={props.app.sessionsByServiceName[service.name]}
                service={service} />)}
    </div>;
})

export default Dashboard;
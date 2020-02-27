Pulling stats from a custom [Carpet](https://github.com/LambdaCraft/fabric-carpet/blob/master/src/main/java/carpet/utils/ServerStatus.java) server and updating players info for a custom [Unmined](https://github.com/LambdaCraft/unmined-wrapper) gen.

Before running, place default `Alex.png` portrait into `/portraits` folder. `/portrains` folder should be placed next to `/<server-name>` generated files.

Example docker run command: 

```bash
docker run -it --rm -v <portraits location>:/portraits -v <unmined gen files>:/generated -e "INTERVAL=60" -e "CARPET=http://host.docker.internal:3141" -e "HEADER_SECRET=CHANGE_ME" vkorn/carpet-stats-lambda
```
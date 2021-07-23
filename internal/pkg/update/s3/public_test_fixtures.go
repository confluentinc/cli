package s3

const ListVersionsPublicFixture = `\
<?xml version="1.0" encoding="utf-8"?>
<ListBucketResult xmlns="http://s3.amazonaws.com/doc/2006-03-01/">
  <Name>confluent.cloud</Name>
  <Prefix>confluent-cli/binaries/</Prefix>
  <Marker></Marker>
  <MaxKeys>1000</MaxKeys>
  <IsTruncated>false</IsTruncated>
  <Contents>
    <Key>confluent-cli/binaries/0.47.0/confluent_0.47.0_checksums.txt</Key>
    <LastModified>2019-03-30T03:47:23.000Z</LastModified>
    <ETag>"6d67d14d2e493c4954f3b9a73c3b7e96"</ETag>
    <Size>469</Size>
    <StorageClass>STANDARD</StorageClass>
  </Contents>
  <Contents>
    <Key>confluent-cli/binaries/0.47.0/confluent_0.47.0_darwin_386</Key>
    <LastModified>2019-03-30T03:47:23.000Z</LastModified>
    <ETag>"a040fb753aa77009330bb7b6e0f805a8"</ETag>
    <Size>17067256</Size>
    <StorageClass>STANDARD</StorageClass>
  </Contents>
  <Contents>
    <Key>confluent-cli/binaries/0.47.0/confluent_0.47.0_darwin_amd64</Key>
    <LastModified>2019-03-30T03:47:23.000Z</LastModified>
    <ETag>"a040fb753aa77009330bb7b6e0f805a8"</ETag>
    <Size>17067256</Size>
    <StorageClass>STANDARD</StorageClass>
  </Contents>
  <Contents>
    <Key>confluent-cli/binaries/0.47.0/confluent_0.47.0_linux_386</Key>
    <LastModified>2019-03-30T03:47:26.000Z</LastModified>
    <ETag>"e8fafdf2321a2fc87b656e403ae47504"</ETag>
    <Size>13308128</Size>
    <StorageClass>STANDARD</StorageClass>
  </Contents>
  <Contents>
    <Key>confluent-cli/binaries/0.47.0/confluent_0.47.0_linux_amd64</Key>
    <LastModified>2019-03-30T03:47:23.000Z</LastModified>
    <ETag>"3c1fb9d89b36d29381b2a45c932ddc5b"</ETag>
    <Size>16110728</Size>
    <StorageClass>STANDARD</StorageClass>
  </Contents>
  <Contents>
    <Key>confluent-cli/binaries/0.47.0/confluent_0.47.0_windows_386.exe</Key>
    <LastModified>2019-03-30T03:47:24.000Z</LastModified>
    <ETag>"f225607534d34dceb7327f733f163623"</ETag>
    <Size>13112832</Size>
    <StorageClass>STANDARD</StorageClass>
  </Contents>
  <Contents>
    <Key>confluent-cli/binaries/0.47.0/confluent_0.47.0_windows_amd64.exe</Key>
    <LastModified>2019-03-30T03:47:23.000Z</LastModified>
    <ETag>"f6144b210abd17f5ef3dd94dfade01b6"</ETag>
    <Size>15363072</Size>
    <StorageClass>STANDARD</StorageClass>
  </Contents>
  <Contents>
    <Key>confluent-cli/binaries/0.48.0/confluent_0.48.0_checksums.txt</Key>
    <LastModified>2019-04-01T22:58:42.000Z</LastModified>
    <ETag>"600585fc2c4b8d4d96fd5f74b2db32df"</ETag>
    <Size>469</Size>
    <StorageClass>STANDARD</StorageClass>
  </Contents>
  <Contents>
    <Key>confluent-cli/binaries/0.48.0/confluent_0.48.0_darwin_amd64</Key>
    <LastModified>2019-04-01T22:58:42.000Z</LastModified>
    <ETag>"73cb792ae2f2997d1a946a44e83d9389"</ETag>
    <Size>17075976</Size>
    <StorageClass>STANDARD</StorageClass>
  </Contents>
  <Contents>
    <Key>confluent-cli/binaries/0.48.0/confluent_0.48.0_linux_386</Key>
    <LastModified>2019-04-01T22:58:45.000Z</LastModified>
    <ETag>"2b7f22ffe01e9482d40f7f2c11bc8755"</ETag>
    <Size>13316320</Size>
    <StorageClass>STANDARD</StorageClass>
  </Contents>
  <Contents>
    <Key>confluent-cli/binaries/0.48.0/confluent_0.48.0_linux_amd64</Key>
    <LastModified>2019-04-01T22:58:42.000Z</LastModified>
    <ETag>"24a970465fb4663993dca032df008b7c"</ETag>
    <Size>16114824</Size>
    <StorageClass>STANDARD</StorageClass>
  </Contents>
  <Contents>
    <Key>confluent-cli/binaries/0.48.0/confluent_0.48.0_windows_386.exe</Key>
    <LastModified>2019-04-01T22:58:43.000Z</LastModified>
    <ETag>"6ed8f95d7962ea36b49f0499c0f26739"</ETag>
    <Size>13117440</Size>
    <StorageClass>STANDARD</StorageClass>
  </Contents>
  <Contents>
    <Key>confluent-cli/binaries/0.48.0/confluent_0.48.0_windows_amd64.exe</Key>
    <LastModified>2019-04-01T22:58:42.000Z</LastModified>
    <ETag>"ed613903fe36012800f9b2a8a318d608"</ETag>
    <Size>15368192</Size>
    <StorageClass>STANDARD</StorageClass>
  </Contents>

</ListBucketResult>`

const ListVersionsPublicFixtureOtherBinaries = `\
<?xml version="1.0" encoding="utf-8"?>
<ListBucketResult xmlns="http://s3.amazonaws.com/doc/2006-03-01/">
  <Name>confluent.cloud</Name>
  <Prefix>confluent-cli/binaries/</Prefix>
  <Marker></Marker>
  <MaxKeys>1000</MaxKeys>
  <IsTruncated>false</IsTruncated>
  <Contents>
    <Key>confluent-cli/binaries/0.42.0/confluent-apikey-plugin_0.42.0_darwin_amd64</Key>
    <LastModified>2019-03-22T21:50:59.000Z</LastModified>
    <ETag>"0407ce7a80fff29882815631947908ab"</ETag>
    <Size>21429624</Size>
    <StorageClass>STANDARD</StorageClass>
  </Contents>
  <Contents>
    <Key>confluent-cli/binaries/0.42.0/confluent-apikey-plugin_0.42.0_linux_386</Key>
    <LastModified>2019-03-22T21:50:50.000Z</LastModified>
    <ETag>"c73e0be5e37f9c671fe9d9ee18742d52"</ETag>
    <Size>16904320</Size>
    <StorageClass>STANDARD</StorageClass>
  </Contents>
  <Contents>
    <Key>confluent-cli/binaries/0.42.0/confluent-apikey-plugin_0.42.0_linux_amd64</Key>
    <LastModified>2019-03-22T21:51:06.000Z</LastModified>
    <ETag>"6476d0c9531ca2ce837a77f89b52924d"</ETag>
    <Size>19648576</Size>
    <StorageClass>STANDARD</StorageClass>
  </Contents>
  <Contents>
    <Key>confluent-cli/binaries/0.42.0/confluent-apikey-plugin_0.42.0_windows_386.exe</Key>
    <LastModified>2019-03-22T21:51:01.000Z</LastModified>
    <ETag>"61474733712934dd201bfa4148e7901c"</ETag>
    <Size>16801792</Size>
    <StorageClass>STANDARD</StorageClass>
  </Contents>
  <Contents>
    <Key>confluent-cli/binaries/0.42.0/confluent-apikey-plugin_0.42.0_windows_amd64.exe</Key>
    <LastModified>2019-03-22T21:51:07.000Z</LastModified>
    <ETag>"f111eed28374eb837d2edb75f6af3fa1"</ETag>
    <Size>19618816</Size>
    <StorageClass>STANDARD</StorageClass>
  </Contents>
  <Contents>
    <Key>confluent-cli/binaries/0.42.0/confluent-kafka-plugin_0.42.0_darwin_amd64</Key>
    <LastModified>2019-03-22T21:50:57.000Z</LastModified>
    <ETag>"99208ee93ac120bfc36f1b61a4b1ba49"</ETag>
    <Size>21498440</Size>
    <StorageClass>STANDARD</StorageClass>
  </Contents>
  <Contents>
    <Key>confluent-cli/binaries/0.42.0/confluent-kafka-plugin_0.42.0_linux_amd64</Key>
    <LastModified>2019-03-22T21:51:02.000Z</LastModified>
    <ETag>"6d8890f05086ada878fc51416ac67efc"</ETag>
    <Size>19710432</Size>
    <StorageClass>STANDARD</StorageClass>
  </Contents>
  <Contents>
    <Key>confluent-cli/binaries/0.42.0/confluent-kafka-plugin_0.42.0_windows_386.exe</Key>
    <LastModified>2019-03-22T21:50:44.000Z</LastModified>
    <ETag>"2fa310f2c2bb3e62088a75ad6d3a6b1f"</ETag>
    <Size>16854016</Size>
    <StorageClass>STANDARD</StorageClass>
  </Contents>
  <Contents>
    <Key>confluent-cli/binaries/0.42.0/confluent-kafka-plugin_0.42.0_windows_amd64.exe</Key>
    <LastModified>2019-03-22T21:50:50.000Z</LastModified>
    <ETag>"2831362a503b2a24d47771d07293c6af"</ETag>
    <Size>19681792</Size>
    <StorageClass>STANDARD</StorageClass>
  </Contents>
  <Contents>
    <Key>confluent-cli/binaries/0.42.0/confluent_0.42.0_checksums.txt</Key>
    <LastModified>2019-03-22T21:50:44.000Z</LastModified>
    <ETag>"9056c6fd3b6350ccbf68e0e7f5686f17"</ETag>
    <Size>2635</Size>
    <StorageClass>STANDARD</StorageClass>
  </Contents>
  <Contents>
    <Key>confluent-cli/binaries/0.42.0/confluent_0.42.0_darwin_amd64</Key>
    <LastModified>2019-03-22T21:50:55.000Z</LastModified>
    <ETag>"c7468c06ca5790a5ac57964ae8cf043a"</ETag>
    <Size>25681544</Size>
    <StorageClass>STANDARD</StorageClass>
  </Contents>
  <Contents>
    <Key>confluent-cli/binaries/0.42.0/confluent_0.42.0_linux_386</Key>
    <LastModified>2019-03-22T21:50:45.000Z</LastModified>
    <ETag>"7df5e40ba2c850517d647fbb834c99c9"</ETag>
    <Size>20158720</Size>
    <StorageClass>STANDARD</StorageClass>
  </Contents>
  <Contents>
    <Key>confluent-cli/binaries/0.42.0/confluent_0.42.0_linux_amd64</Key>
    <LastModified>2019-03-22T21:51:04.000Z</LastModified>
    <ETag>"0422ee9fb50e4d691afa5214502b6b0b"</ETag>
    <Size>24026792</Size>
    <StorageClass>STANDARD</StorageClass>
  </Contents>
  <Contents>
    <Key>confluent-cli/binaries/0.42.0/confluent_0.42.0_windows_386.exe</Key>
    <LastModified>2019-03-22T21:51:01.000Z</LastModified>
    <ETag>"d0a3a3029e01fd47f24ab982328f2d7f"</ETag>
    <Size>20067328</Size>
    <StorageClass>STANDARD</StorageClass>
  </Contents>
  <Contents>
    <Key>confluent-cli/binaries/0.42.0/confluent_0.42.0_windows_amd64.exe</Key>
    <LastModified>2019-03-22T21:50:51.000Z</LastModified>
    <ETag>"cae9fe523062810bf3b6eb06a6ec599c"</ETag>
    <Size>23444992</Size>
    <StorageClass>STANDARD</StorageClass>
  </Contents>


  <Contents>
    <Key>confluent-cli/binaries/0.44.0/other_0.44.0_darwin_amd64</Key>
    <LastModified>2019-03-29T20:32:38.000Z</LastModified>
    <ETag>"42ceaf9337d08be81d625f6ede2d62c7"</ETag>
    <Size>17067256</Size>
    <StorageClass>STANDARD</StorageClass>
  </Contents>
  <Contents>
    <Key>confluent-cli/binaries/0.44.0/other_0.44.0_linux_386</Key>
    <LastModified>2019-03-29T20:32:40.000Z</LastModified>
    <ETag>"a46b105d553f1dceaf344f199e1f47d9"</ETag>
    <Size>13308128</Size>
    <StorageClass>STANDARD</StorageClass>
  </Contents>
  <Contents>
    <Key>confluent-cli/binaries/0.44.0/other_0.44.0_linux_amd64</Key>
    <LastModified>2019-03-29T20:32:39.000Z</LastModified>
    <ETag>"ac08671d22bcf68cbe70fb1d4de0f38e"</ETag>
    <Size>16110728</Size>
    <StorageClass>STANDARD</StorageClass>
  </Contents>
  <Contents>
    <Key>confluent-cli/binaries/0.44.0/other_0.44.0_windows_386.exe</Key>
    <LastModified>2019-03-29T20:32:36.000Z</LastModified>
    <ETag>"211fa1769cdc2748bc530768c2033ffb"</ETag>
    <Size>13112832</Size>
    <StorageClass>STANDARD</StorageClass>
  </Contents>
  <Contents>
    <Key>confluent-cli/binaries/0.44.0/other_0.44.0_windows_amd64.exe</Key>
    <LastModified>2019-03-29T20:32:40.000Z</LastModified>
    <ETag>"c0679052427a70fb9ba5ac8cdcee59de"</ETag>
    <Size>15363072</Size>
    <StorageClass>STANDARD</StorageClass>
  </Contents>


</ListBucketResult>`

const ListVersionsPublicFixtureDirtyVersions = `\
<?xml version="1.0" encoding="utf-8"?>
<ListBucketResult xmlns="http://s3.amazonaws.com/doc/2006-03-01/">
  <Name>confluent.cloud</Name>
  <Prefix>confluent-cli/binaries/</Prefix>
  <Marker></Marker>
  <MaxKeys>1000</MaxKeys>
  <IsTruncated>false</IsTruncated>
  <Contents>
    <Key>confluent-cli/binaries/0.44.0/confluent_0.44.0_darwin_amd64</Key>
    <LastModified>2019-03-29T20:32:38.000Z</LastModified>
    <ETag>"42ceaf9337d08be81d625f6ede2d62c7"</ETag>
    <Size>17067256</Size>
    <StorageClass>STANDARD</StorageClass>
  </Contents>
  <Contents>
    <Key>confluent-cli/binaries/0.44.0/confluent_0.44.0-dirty-cody_darwin_amd64</Key>
    <LastModified>2019-03-29T20:32:38.000Z</LastModified>
    <ETag>"42ceaf9337d08be81d625f6ede2d62c7"</ETag>
    <Size>17067256</Size>
    <StorageClass>STANDARD</StorageClass>
  </Contents>
  <Contents>
    <Key>confluent-cli/binaries/0.44.0/confluent_0.44.0-SNAPSHOT_darwin_amd64</Key>
    <LastModified>2019-03-29T20:32:38.000Z</LastModified>
    <ETag>"42ceaf9337d08be81d625f6ede2d62c7"</ETag>
    <Size>17067256</Size>
    <StorageClass>STANDARD</StorageClass>
  </Contents>
</ListBucketResult>`

const ListVersionsPublicFixtureUnsortedVersions = `\
<?xml version="1.0" encoding="utf-8"?>
<ListBucketResult xmlns="http://s3.amazonaws.com/doc/2006-03-01/">
  <Name>confluent.cloud</Name>
  <Prefix>confluent-cli/binaries/</Prefix>
  <Marker></Marker>
  <MaxKeys>1000</MaxKeys>
  <IsTruncated>false</IsTruncated>
  <Contents>
    <Key>confluent-cli/binaries/0.43.0/confluent_0.43.0_darwin_amd64</Key>
    <LastModified>2019-03-29T20:32:38.000Z</LastModified>
    <ETag>"42ceaf9337d08be81d625f6ede2d62c7"</ETag>
    <Size>17067256</Size>
    <StorageClass>STANDARD</StorageClass>
  </Contents>
  <Contents>
    <Key>confluent-cli/binaries/0.42.0/confluent_0.42.0_darwin_amd64</Key>
    <LastModified>2019-03-29T20:32:38.000Z</LastModified>
    <ETag>"42ceaf9337d08be81d625f6ede2d62c7"</ETag>
    <Size>17067256</Size>
    <StorageClass>STANDARD</StorageClass>
  </Contents>
  <Contents>
    <Key>confluent-cli/binaries/0.42.1/confluent_0.42.1_darwin_amd64</Key>
    <LastModified>2019-03-29T20:32:38.000Z</LastModified>
    <ETag>"42ceaf9337d08be81d625f6ede2d62c7"</ETag>
    <Size>17067256</Size>
    <StorageClass>STANDARD</StorageClass>
  </Contents>
</ListBucketResult>`

const ListVersionsPublicFixtureNonSemver = `\
<?xml version="1.0" encoding="utf-8"?>
<ListBucketResult xmlns="http://s3.amazonaws.com/doc/2006-03-01/">
  <Name>confluent.cloud</Name>
  <Prefix>confluent-cli/binaries/</Prefix>
  <Marker></Marker>
  <MaxKeys>1000</MaxKeys>
  <IsTruncated>false</IsTruncated>
  <Contents>
    <Key>confluent-cli/binaries/v1beta1/confluent_v1beta1_darwin_amd64</Key>
    <LastModified>2019-03-29T20:32:38.000Z</LastModified>
    <ETag>"42ceaf9337d08be81d625f6ede2d62c7"</ETag>
    <Size>17067256</Size>
    <StorageClass>STANDARD</StorageClass>
  </Contents>
</ListBucketResult>`

const ListVersionsPublicFixtureInvalidNames = `\
<?xml version="1.0" encoding="utf-8"?>
<ListBucketResult xmlns="http://s3.amazonaws.com/doc/2006-03-01/">
  <Name>confluent.cloud</Name>
  <Prefix>confluent-cli/binaries/</Prefix>
  <Marker></Marker>
  <MaxKeys>1000</MaxKeys>
  <IsTruncated>false</IsTruncated>
  <Contents>
    <Key>confluent-cli/binaries/0.42.0/confluent</Key>
    <LastModified>2019-03-29T20:32:38.000Z</LastModified>
    <ETag>"42ceaf9337d08be81d625f6ede2d62c7"</ETag>
    <Size>17067256</Size>
    <StorageClass>STANDARD</StorageClass>
  </Contents>
</ListBucketResult>`

const ListVersionsPublicFixtureInvalidPrefix = `\
<?xml version="1.0" encoding="utf-8"?>
<ListBucketResult xmlns="http://s3.amazonaws.com/doc/2006-03-01/">
  <Name>confluent.cloud</Name>
  <Prefix>confluent-cli/binaries/</Prefix>
  <Marker></Marker>
  <MaxKeys>1000</MaxKeys>
  <IsTruncated>false</IsTruncated>
  <Contents>
    <Key>confluent-cli/binaries/0.43.0/confluent_0.43.0_darwin_amd64</Key>
    <LastModified>2019-03-29T20:32:38.000Z</LastModified>
    <ETag>"42ceaf9337d08be81d625f6ede2d62c7"</ETag>
    <Size>17067256</Size>
    <StorageClass>STANDARD</StorageClass>
  </Contents>
</ListBucketResult>`

const ListReleaseNotesVersionsPublicFixture = `\
<?xml version="1.0" encoding="utf-8"?>
<ListBucketResult xmlns="http://s3.amazonaws.com/doc/2006-03-01/">
  <Name>confluent.cloud</Name>
  <Prefix>confluent-cli/binaries/</Prefix>
  <Marker></Marker>
  <MaxKeys>1000</MaxKeys>
  <IsTruncated>false</IsTruncated>
  <Contents>
    <Key>confluent-cli/release-notes/0.0.0/release-notes.rst</Key>
    <LastModified>2019-03-30T03:47:23.000Z</LastModified>
    <ETag>"6d67d14d2e493c4954f3b9a73c3b7e96"</ETag>
    <Size>469</Size>
    <StorageClass>STANDARD</StorageClass>
  </Contents>
  <Contents>
    <Key>confluent-cli/release-notes/0.1.0/release-notes.rst</Key>
    <LastModified>2019-03-30T03:47:23.000Z</LastModified>
    <ETag>"a040fb753aa77009330bb7b6e0f805a8"</ETag>
    <Size>469</Size>
    <StorageClass>STANDARD</StorageClass>
  </Contents>

</ListBucketResult>`

const ListReleaseNotesVersionsInvalidFiles = `\
<?xml version="1.0" encoding="utf-8"?>
<ListBucketResult xmlns="http://s3.amazonaws.com/doc/2006-03-01/">
  <Name>confluent.cloud</Name>
  <Prefix>confluent-cli/binaries/</Prefix>
  <Marker></Marker>
  <MaxKeys>1000</MaxKeys>
  <IsTruncated>false</IsTruncated>
  <Contents>
    <Key>confluent-cli/release-notes/0.47.0/bababa.rst</Key>
    <LastModified>2019-03-30T03:47:23.000Z</LastModified>
    <ETag>"6d67d14d2e493c4954f3b9a73c3b7e96"</ETag>
    <Size>469</Size>
    <StorageClass>STANDARD</StorageClass>
  </Contents>
  <Contents>
    <Key>confluent-cli/release-notes/0.48.0/blob</Key>
    <LastModified>2019-03-30T03:47:23.000Z</LastModified>
    <ETag>"a040fb753aa77009330bb7b6e0f805a8"</ETag>
    <Size>469</Size>
    <StorageClass>STANDARD</StorageClass>
  </Contents>

</ListBucketResult>`

const ListReleaseNotesVersionsExcludeInvalidFiles = `\
<?xml version="1.0" encoding="utf-8"?>
<ListBucketResult xmlns="http://s3.amazonaws.com/doc/2006-03-01/">
  <Name>confluent.cloud</Name>
  <Prefix>confluent-cli/binaries/</Prefix>
  <Marker></Marker>
  <MaxKeys>1000</MaxKeys>
  <IsTruncated>false</IsTruncated>
  <Contents>
    <Key>confluent-cli/release-notes/0.47.0/release-notes.rst</Key>
    <LastModified>2019-03-30T03:47:23.000Z</LastModified>
    <ETag>"6d67d14d2e493c4954f3b9a73c3b7e96"</ETag>
    <Size>469</Size>
    <StorageClass>STANDARD</StorageClass>
  </Contents>
  <Contents>
    <Key>confluent-cli/release-notes/0.48.0/blobblab</Key>
    <LastModified>2019-03-30T03:47:23.000Z</LastModified>
    <ETag>"a040fb753aa77009330bb7b6e0f805a8"</ETag>
    <Size>469</Size>
    <StorageClass>STANDARD</StorageClass>
  </Contents>

</ListBucketResult>`

const ListReleaseNotesVersionsPublicFixtureUnsortedVersions = `\
<?xml version="1.0" encoding="utf-8"?>
<ListBucketResult xmlns="http://s3.amazonaws.com/doc/2006-03-01/">
  <Name>confluent.cloud</Name>
  <Prefix>confluent-cli/binaries/</Prefix>
  <Marker></Marker>
  <MaxKeys>1000</MaxKeys>
  <IsTruncated>false</IsTruncated>
  <Contents>
    <Key>confluent-cli/release-notes/0.43.0/release-notes.rst</Key>
    <LastModified>2019-03-29T20:32:38.000Z</LastModified>
    <ETag>"42ceaf9337d08be81d625f6ede2d62c7"</ETag>
    <Size>17067256</Size>
    <StorageClass>STANDARD</StorageClass>
  </Contents>
  <Contents>
    <Key>confluent-cli/release-notes/0.42.0/release-notes.rst</Key>
    <LastModified>2019-03-29T20:32:38.000Z</LastModified>
    <ETag>"42ceaf9337d08be81d625f6ede2d62c7"</ETag>
    <Size>17067256</Size>
    <StorageClass>STANDARD</StorageClass>
  </Contents>
  <Contents>
    <Key>confluent-cli/release-notes/0.42.1/release-notes.rst</Key>
    <LastModified>2019-03-29T20:32:38.000Z</LastModified>
    <ETag>"42ceaf9337d08be81d625f6ede2d62c7"</ETag>
    <Size>17067256</Size>
    <StorageClass>STANDARD</StorageClass>
  </Contents>
</ListBucketResult>`

const ListReleaseNotesVersionsPublicFixtureNonSemver = `\
<?xml version="1.0" encoding="utf-8"?>
<ListBucketResult xmlns="http://s3.amazonaws.com/doc/2006-03-01/">
  <Name>confluent.cloud</Name>
  <Prefix>confluent-cli/binaries/</Prefix>
  <Marker></Marker>
  <MaxKeys>1000</MaxKeys>
  <IsTruncated>false</IsTruncated>
  <Contents>
    <Key>confluent-cli/release-notes/v1beta1/release-notes.rst</Key>
    <LastModified>2019-03-29T20:32:38.000Z</LastModified>
    <ETag>"42ceaf9337d08be81d625f6ede2d62c7"</ETag>
    <Size>17067256</Size>
    <StorageClass>STANDARD</StorageClass>
  </Contents>
</ListBucketResult>`

const ReleaseNotesFileV0470 = `
===================================
Confluent CLI v0.47.0 Release Notes
===================================
`
